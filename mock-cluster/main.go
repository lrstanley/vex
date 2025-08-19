// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-faker/faker/v4"
	vapi "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/userpass"
	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
)

var (
	cli    = &Flags{}
	logger = slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
			AddSource:  true,
		}),
	)
)

type Flags struct {
	Init struct {
		Image           string `default:"hashicorp/vault" help:"image to use for the mock cluster"`
		Version         string `default:"latest" help:"version of the mock cluster"`
		NumNodes        int    `default:"3" help:"number of nodes in the mock cluster"`
		UnsealKeys      int    `default:"1" help:"number of unseal keys to generate"`
		UnsealThreshold int    `default:"1" help:"number of unseal keys required to unseal the cluster"`
	} `cmd:"" help:"initialize the mock cluster"`
	Start     struct{} `cmd:"" help:"start the mock cluster, and bootstrap if not already bootstrapped"`
	Bootstrap struct {
		Force bool `help:"force bootstrap even if already bootstrapped"`
	} `cmd:"" help:"bootstrap the mock cluster"`
	Stop struct {
		Timeout time.Duration `help:"timeout for the stop command"`
	} `cmd:"" help:"stop the mock cluster"`
	RM  struct{} `cmd:"" help:"remove the mock cluster"`
	Env struct {
		Node int    `default:"1" help:"node to print the env vars for"`
		ID   string `default:"token" help:"user to print the env vars for, or 'token' to use the root token"`
	} `cmd:"" help:"print the env vars for a user"`
}

func main() {
	LoadConfig()

	cctx := kong.Parse(
		cli,
		kong.Name("mock-cluster"),
		kong.Description("Mock cluster for testing"),
		kong.UsageOnError(),
	)

	ctx := context.Background()

	switch cctx.Command() {
	case "init":
		err := create(ctx)
		if err != nil {
			logger.Error("failed to initialize mock cluster (but don't start)", "error", err)
			os.Exit(1)
		}
		return
	case "start":
		if cli.Init.NumNodes < 3 {
			logger.Error("number of nodes must be at least 3")
			os.Exit(1)
		}

		err := start(ctx)
		if err != nil {
			logger.Error("failed to start mock cluster", "error", err)
			os.Exit(1)
		}
		return
	case "bootstrap":
		err := bootstrap(ctx)
		if err != nil {
			logger.Error("failed to bootstrap mock cluster", "error", err)
			os.Exit(1)
		}
		return
	case "stop":
		err := stop(ctx)
		if err != nil {
			logger.Error("failed to stop mock cluster", "error", err)
			os.Exit(1)
		}
		return
	case "rm":
		err := rm(ctx)
		if err != nil {
			logger.Error("failed to remove mock cluster", "error", err)
			os.Exit(1)
		}
		return
	case "env":
		fmt.Println("VAULT_ADDR=http://127.0.0.1:820" + strconv.Itoa(cli.Env.Node-1)) //nolint:forbidigo
		if cli.Env.ID == "token" {
			fmt.Println("VAULT_TOKEN=" + config.RootToken) //nolint:forbidigo
		} else {
			vault := NewVaultClient(ctx, 1)
			up, err := userpass.NewUserpassAuth(cli.Env.ID, &userpass.Password{
				FromString: config.Users[cli.Env.ID],
			})
			if err != nil {
				logger.Error("failed to create userpass auth", "error", err)
				os.Exit(1)
			}
			_, err = vault.Auth().Login(ctx, up)
			if err != nil {
				logger.Error("failed to login", "error", err)
				os.Exit(1)
			}
			fmt.Println("VAULT_TOKEN=" + vault.Token()) //nolint:forbidigo
		}
		return
	}
}

func create(ctx context.Context) error {
	dkr := NewDockerClient(ctx)
	defer dkr.Close()

	logger.InfoContext(ctx, "creating mock cluster", "num_nodes", cli.Init.NumNodes)

	for i := range cli.Init.NumNodes {
		_, err := DockerGetNode(ctx, dkr, i+1)
		if err == nil {
			logger.InfoContext(ctx, "node already exists", "node", i+1)
			continue
		}
		err = DockerCreateNode(ctx, dkr, i+1)
		if err != nil {
			return fmt.Errorf("failed to create node: %w", err)
		}
	}

	return nil
}

func start(ctx context.Context) error {
	dkr := NewDockerClient(ctx)
	defer dkr.Close()

	var err error

	ps := DockerGetContainers(ctx, dkr)
	if len(ps) == 0 {
		err = create(ctx)
		if err != nil {
			return err
		}
		ps = DockerGetContainers(ctx, dkr)
	}

	if len(ps) != cli.Init.NumNodes {
		return fmt.Errorf("expected %d nodes, got %d (recreate the cluster)", cli.Init.NumNodes, len(ps))
	}

	logger.InfoContext(ctx, "starting mock cluster", "num_nodes", len(ps))

	err = DockerClusterStart(ctx, dkr, 20*time.Second)
	if err != nil {
		return fmt.Errorf("failed to start cluster: %w", err)
	}

	err = WaitVaultClusterAvailable(ctx, 60*time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for vault cluster to be available: %w", err)
	}

	err = WaitVaultInitialize(ctx, 60*time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for vault cluster to be initialized: %w", err)
	}

	var eg errgroup.Group

	for i := range cli.Init.NumNodes {
		eg.Go(func() error {
			return WaitVaultUnseal(ctx, 120*time.Second, i+1)
		})
	}

	err = eg.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for vault cluster to be unsealed: %w", err)
	}

	for i := range cli.Init.NumNodes {
		err = WaitVaultHealthy(ctx, 120*time.Second, i+1)
		if err != nil {
			return fmt.Errorf("failed to wait for vault cluster to be healthy: %w", err)
		}
	}

	return bootstrap(ctx)
}

func bootstrap(ctx context.Context) error { //nolint:funlen,unparam
	var wg sync.WaitGroup

	if !config.BootstrappedPolicies || cli.Bootstrap.Force {
		BootstrapVaultPolicy(ctx, "sudo-policy", `
# Root policy with sudo permissions
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}`)

		BootstrapVaultPolicy(ctx, "admin-policy", `
# Admin policy for general administration
path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/policies/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv-v1-*/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv-v2-*/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}`)

		BootstrapVaultPolicy(ctx, "ops-policy", `
# Operations policy ${i}
path "sys/health" {
  capabilities = ["read"]
}

path "kv-v1-*/*" {
  capabilities = ["read", "list"]
}

path "kv-v2-*/*" {
  capabilities = ["read", "list"]
}`)

		BootstrapVaultPolicy(ctx, "dev-policy", `
# Developer policy
path "kv-v1-1/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv-v2-1/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}`)

		BootstrapVaultPolicy(ctx, "readonly-policy", `
# Readonly policy
path "*" {
  capabilities = ["read", "list"]
}`)

		config.BootstrappedPolicies = true
		SaveConfig()
	}

	if !config.BootstrappedAuthEngines || cli.Bootstrap.Force {
		BootstrapVaultAuthEngines(ctx)
		config.BootstrappedAuthEngines = true
		SaveConfig()
	}

	if !config.BootstrappedUsers || cli.Bootstrap.Force {
		BootstrapVaultUser(ctx, "root", "sudo-policy")
		for range 10 {
			wg.Go(func() {
				BootstrapVaultUser(ctx, "admin-"+strings.ToLower(faker.FirstName()), "admin-policy")
				BootstrapVaultUser(ctx, "ops-"+strings.ToLower(faker.FirstName()), "ops-policy")
				BootstrapVaultUser(ctx, "dev-"+strings.ToLower(faker.FirstName()), "dev-policy")
				BootstrapVaultUser(ctx, "readonly-"+strings.ToLower(faker.FirstName()), "readonly-policy")
			})
		}

		wg.Wait()
		config.BootstrappedUsers = true
		SaveConfig()
	}

	if !config.BootstrappedMounts || cli.Bootstrap.Force {
		for i := range 10 {
			wg.Go(func() {
				BootstrapVaultMount(ctx, "kv-v1-"+strconv.Itoa(i+1), vapi.MountInput{
					Type:        "kv",
					Description: "mock kv v1 secret engine",
					Options: map[string]string{
						"version": "1",
					},
				})
				BootstrapVaultMount(ctx, "kv-v2-"+strconv.Itoa(i+1), vapi.MountInput{
					Type:        "kv",
					Description: "mock kv v2 secret engine",
					Options:     map[string]string{"version": "2"},
				})
			})
		}
		BootstrapVaultMount(ctx, "transit", vapi.MountInput{
			Type:        "transit",
			Description: "mock transit secret engine",
		})
		BootstrapVaultMount(ctx, "ssh", vapi.MountInput{
			Type:        "ssh",
			Description: "mock ssh secret engine",
		})
		BootstrapVaultMount(ctx, "pki", vapi.MountInput{
			Type:        "pki",
			Description: "mock pki secret engine",
			Config: vapi.MountConfigInput{
				MaxLeaseTTL:     "3650d",
				DefaultLeaseTTL: "365d",
			},
		})

		wg.Wait()
		config.BootstrappedMounts = true
		SaveConfig()
	}

	if !config.BootstrappedKVSecrets || cli.Bootstrap.Force {
		certs := GetHostCertificates("github.com:443")

		values := map[string]any{
			"key1-password": faker.Password(),
			"key2-password": faker.Password(),
			"longer-key":    faker.Paragraph(),
			"single-cert":   CertChainString(certs[0]),
			"cert-chain":    CertChainString(certs...),
		}

		values2 := maps.Clone(values)
		values2["foo"] = map[string]any{
			"bar": map[string]any{
				"baz": faker.Password(),
			},
		}

		for i := range 10 {
			for range 200 {
				wg.Go(func() {
					BootstrapVaultKVSecrets(
						ctx,
						"kv-v1-"+strconv.Itoa(i+1)+"/"+GenerateSlug(5, "/"),
						values,
					)
				})

				wg.Go(func() {
					key := GenerateSlug(5, "/")
					BootstrapVaultKVSecrets(
						ctx,
						"kv-v2-"+strconv.Itoa(i+1)+"/data/"+key,
						map[string]any{
							"data":    values2,
							"options": map[string]any{},
						},
					)
				})
			}
		}
		wg.Wait()
		config.BootstrappedKVSecrets = true
		SaveConfig()
	}

	return nil
}

func stop(ctx context.Context) error {
	dkr := NewDockerClient(ctx)
	defer dkr.Close()

	logger.InfoContext(ctx, "stopping mock cluster")

	return DockerClusterStop(ctx, dkr, 20*time.Second, false)
}

func rm(ctx context.Context) error {
	dkr := NewDockerClient(ctx)
	defer dkr.Close()

	logger.InfoContext(ctx, "removing mock cluster")

	err := DockerClusterRemove(ctx, dkr, 20*time.Second)
	if err != nil {
		return fmt.Errorf("failed to remove cluster: %w", err)
	}

	err = DockerDeleteNetwork(ctx, dkr)
	if err != nil {
		return fmt.Errorf("failed to delete network: %w", err)
	}

	RemoveConfig()
	return nil
}
