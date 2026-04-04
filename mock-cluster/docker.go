// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	MockClusterLabel = "com.lrstanley.vex.mock-cluster"
	MockNodeLabel    = "com.lrstanley.vex.mock-cluster.node"
)

func natPortBindingToNetwork(b nat.PortBinding) network.PortBinding {
	var addr netip.Addr
	if b.HostIP != "" {
		var err error
		addr, err = netip.ParseAddr(b.HostIP)
		if err != nil {
			addr = netip.Addr{}
		}
	}
	return network.PortBinding{HostIP: addr, HostPort: b.HostPort}
}

func NewDockerClient(ctx context.Context) *client.Client {
	dkr := Must(client.New(client.FromEnv))
	info := Must(dkr.Info(ctx, client.InfoOptions{}))
	logger.InfoContext(ctx,
		"docker client created",
		"version", info.Info.ServerVersion,
		"os", info.Info.OperatingSystem,
		"arch", info.Info.Architecture,
	)
	return dkr
}

func DockerGetContainers(ctx context.Context, dkr *client.Client) []container.Summary {
	res := Must(dkr.ContainerList(ctx, client.ContainerListOptions{
		All:   true,
		Limit: 100,
		Filters: client.Filters{}.
			Add("name", "vex-vault-").
			Add("label", MockClusterLabel+"=true"),
	}))
	return res.Items
}

func DockerPullImage(ctx context.Context, dkr *client.Client, img string) error {
	logger.InfoContext(ctx, "pulling image", "image", img)
	f, err := dkr.ImagePull(ctx, img, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	_, err = io.Copy(io.Discard, f)
	if err != nil {
		return fmt.Errorf("failed to copy image: %w", err)
	}
	return nil
}

func DockerClusterStart(ctx context.Context, dkr *client.Client, timeout time.Duration) error {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := DockerPullImage(tctx, dkr, cli.Flags.Init.VaultImage+":"+cli.Flags.Init.VaultVersion)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	containers := DockerGetContainers(tctx, dkr)
	if len(containers) == 0 {
		return errors.New("no containers found")
	}

	for _, c := range containers {
		if c.State == container.StateRunning {
			logger.InfoContext(ctx, "container already running", "id", c.ID, "name", strings.Join(c.Names, ", "))
			continue
		}

		logger.InfoContext(ctx, "starting container", "id", c.ID, "name", strings.Join(c.Names, ", "))

		_, err = dkr.ContainerStart(tctx, c.ID, client.ContainerStartOptions{})
		if err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}
	}

	logger.InfoContext(ctx, "cluster started", "count", len(containers))
	return nil
}

func DockerClusterStop(ctx context.Context, dkr *client.Client, timeout time.Duration, kill bool) error {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var stopTimeout *int
	if kill {
		stopTimeout = Ptr(0)
	} else {
		stopTimeout = Ptr(2)
	}

	containers := DockerGetContainers(tctx, dkr)
	if len(containers) == 0 {
		return nil
	}

	for _, c := range containers {
		if c.State == container.StateExited {
			logger.InfoContext(ctx, "container already stopped", "id", c.ID, "name", strings.Join(c.Names, ", "))
			continue
		}

		logger.InfoContext(ctx, "stopping container", "id", c.ID, "name", strings.Join(c.Names, ", "))

		_, err := dkr.ContainerStop(tctx, c.ID, client.ContainerStopOptions{
			Timeout: stopTimeout,
		})
		if err != nil {
			return fmt.Errorf("failed to stop container: %w", err)
		}
	}

	logger.InfoContext(ctx, "cluster stopped", "count", len(containers))
	return nil
}

func DockerClusterRemove(ctx context.Context, dkr *client.Client, timeout time.Duration) error {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for _, c := range DockerGetContainers(tctx, dkr) {
		logger.InfoContext(ctx, "removing container", "id", c.ID, "name", strings.Join(c.Names, ", "))

		_, err := dkr.ContainerRemove(tctx, c.ID, client.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
		if err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	volumes, err := dkr.VolumeList(tctx, client.VolumeListOptions{
		Filters: client.Filters{}.Add("label", MockClusterLabel+"=true"),
	})
	if err != nil {
		return fmt.Errorf("failed to list volumes: %w", err)
	}

	for _, v := range volumes.Items {
		logger.InfoContext(ctx, "removing volume", "name", v.Name)
		_, err = dkr.VolumeRemove(tctx, v.Name, client.VolumeRemoveOptions{Force: true})
		if err != nil {
			return fmt.Errorf("failed to remove volume: %w", err)
		}
	}

	logger.InfoContext(ctx, "cluster removed")
	return nil
}

func DockerGetNetwork(ctx context.Context, dkr *client.Client) (*network.Summary, error) {
	nets, err := dkr.NetworkList(ctx, client.NetworkListOptions{
		Filters: client.Filters{}.Add("label", MockClusterLabel+"=true"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	if len(nets.Items) > 1 {
		names := make([]string, 0, len(nets.Items))
		for _, net := range nets.Items {
			names = append(names, net.Name)
		}
		return nil, fmt.Errorf("multiple networks found: %s", strings.Join(names, ", "))
	}

	if len(nets.Items) == 0 {
		return nil, errors.New("no network found")
	}

	return &nets.Items[0], nil
}

func DockerCreateNetwork(ctx context.Context, dkr *client.Client) *network.Summary {
	logger.InfoContext(ctx, "creating network")
	_, err := dkr.NetworkCreate(ctx, "vex-vault", client.NetworkCreateOptions{
		Driver:     "bridge",
		Scope:      "local",
		EnableIPv4: Ptr(true),
		EnableIPv6: Ptr(false),
		Labels: map[string]string{
			MockClusterLabel: "true",
		},
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to create network", "error", err)
		os.Exit(1)
	}

	net, err := DockerGetNetwork(ctx, dkr)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get network", "error", err)
		os.Exit(1)
	}
	return net
}

func DockerGetOrCreateNetwork(ctx context.Context, dkr *client.Client) *network.Summary {
	net, err := DockerGetNetwork(ctx, dkr)
	if err != nil {
		return DockerCreateNetwork(ctx, dkr)
	}
	return net
}

func DockerDeleteNetwork(ctx context.Context, dkr *client.Client) error {
	logger.InfoContext(ctx, "deleting network")
	net, err := DockerGetNetwork(ctx, dkr)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	_, err = dkr.NetworkRemove(ctx, net.ID, client.NetworkRemoveOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete network: %w", err)
	}

	logger.InfoContext(ctx, "network deleted")
	return nil
}

func DockerGetNode(ctx context.Context, dkr *client.Client, node int) (*container.Summary, error) {
	containers := DockerGetContainers(ctx, dkr)
	if len(containers) == 0 {
		return nil, errors.New("no containers found")
	}
	for _, c := range containers {
		if c.Labels[MockNodeLabel] == strconv.Itoa(node) {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("node %d not found", node)
}

func DockerCreateNode(ctx context.Context, dkr *client.Client, node int) error {
	logger.InfoContext(ctx, "creating container for node", "node", node)

	vol, err := dkr.VolumeCreate(ctx, client.VolumeCreateOptions{
		Name:   "vex-vault-" + strconv.Itoa(node),
		Driver: "local",
		Labels: map[string]string{
			MockClusterLabel: "true",
			MockNodeLabel:    strconv.Itoa(node),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create data volume: %w", err)
	}

	net := DockerGetOrCreateNetwork(ctx, dkr)

	natExposed, natBindings, err := nat.ParsePortSpecs([]string{
		fmt.Sprintf("%d:8200/tcp", 8200+node-1),
		"8201/tcp",
	})
	if err != nil {
		return fmt.Errorf("failed to parse port specs: %w", err)
	}

	exposedPorts := network.PortSet{}
	for np := range natExposed {
		p, perr := network.ParsePort(string(np))
		if perr != nil {
			return fmt.Errorf("parse container port: %w", perr)
		}
		exposedPorts[p] = struct{}{}
	}
	portBindings := network.PortMap{}
	for np, binds := range natBindings {
		p, perr := network.ParsePort(string(np))
		if perr != nil {
			return fmt.Errorf("parse container port: %w", perr)
		}
		if len(binds) == 0 {
			continue
		}
		nb := make([]network.PortBinding, len(binds))
		for i, b := range binds {
			nb[i] = natPortBindingToNetwork(b)
		}
		portBindings[p] = nb
	}

	alias := "vault" + strconv.Itoa(node)

	config := ExecTmpl("config.hcl.gotmpl", map[string]any{
		"NodeHostname": alias,
		"NodeID":       node,
		"NumNodes":     cli.Flags.Init.NumNodes,
	})

	resp, err := dkr.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image: cli.Flags.Init.VaultImage + ":" + cli.Flags.Init.VaultVersion,
			Env:   []string{},
			Labels: map[string]string{
				MockClusterLabel: "true",
				MockNodeLabel:    strconv.Itoa(node),
			},
			Hostname:     alias,
			User:         "root",
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
			Entrypoint: []string{
				"/bin/sh",
				"-c",
				fmt.Sprintf(
					"chmod 777 /vault/data && mkdir -p /vault/config && echo -e %q > /vault/config/config.hcl && /usr/local/bin/docker-entrypoint.sh server",
					config,
				),
			},
			ExposedPorts: exposedPorts,
		},
		HostConfig: &container.HostConfig{
			NetworkMode:  container.NetworkMode(net.ID),
			PortBindings: portBindings,
			RestartPolicy: container.RestartPolicy{
				Name: container.RestartPolicyUnlessStopped,
			},
			CapAdd: []string{"CAP_IPC_LOCK"},
			Mounts: []mount.Mount{
				{
					Type:        mount.TypeVolume,
					Source:      vol.Volume.Name,
					Target:      "/vault/data",
					ReadOnly:    false,
					Consistency: mount.ConsistencyDefault,
				},
			},
		},
		NetworkingConfig: &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				net.ID: {
					Aliases:  []string{alias},
					DNSNames: []string{alias},
				},
			},
		},
		Platform: &ocispec.Platform{},
		Name:     "vex-vault-" + strconv.Itoa(node),
	})
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	logger.InfoContext(ctx, "container created", "id", resp.ID)
	return nil
}
