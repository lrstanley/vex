// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	vapi "github.com/hashicorp/vault/api"
	"golang.org/x/time/rate"
)

func NewVaultClient(ctx context.Context, node int) *vapi.Client {
	addr := fmt.Sprintf("http://localhost:%d", 8200+node-1)

	cfg := &vapi.Config{
		HttpClient:       cleanhttp.DefaultPooledClient(),
		Backoff:          retryablehttp.LinearJitterBackoff,
		Address:          addr,
		MaxRetries:       5,
		DisableRedirects: true,
		MinRetryWait:     3 * time.Second,
		MaxRetryWait:     10 * time.Second,
		Timeout:          60 * time.Second,
		Limiter:          rate.NewLimiter(rate.Every(100*time.Millisecond), 1),
	}

	c, err := vapi.NewClient(cfg)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create vault client for node", "node", node, "error", err)
		os.Exit(1)
	}

	if config.RootToken != "" {
		c.SetToken(config.RootToken)
	} else {
		c.SetToken("")
	}

	return c
}

func WaitVaultClusterAvailable(ctx context.Context, timeout time.Duration) error {
	// Do a basic port check.
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for i := range cli.Flags.Init.NumNodes {
		success := false
		for attempt := range 30 {
			if tctx.Err() != nil {
				return tctx.Err()
			}

			logger.InfoContext(ctx, "checking if vault is online", "node", i+1, "attempt", attempt)
			conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(8200+i), 1*time.Second) //nolint:noctx
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			_ = conn.Close()
			logger.InfoContext(ctx, "vault is online", "node", i+1)
			success = true
			break
		}
		if !success {
			return fmt.Errorf("failed to connect to vault node %d in time", i+1)
		}
	}

	return nil
}

func WaitVaultInitialize(ctx context.Context, timeout time.Duration) error {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// We'll initialize the first node.

	vault := NewVaultClient(ctx, 1)

	hasInitialized, err := vault.Sys().InitStatusWithContext(tctx)
	if err != nil {
		return fmt.Errorf("failed to get init status: %w", err)
	}

	if hasInitialized {
		logger.InfoContext(ctx, "vault is already initialized")
		return nil
	}

	logger.InfoContext(ctx, "initializing vault cluster", "node", 1)

	resp, err := vault.Sys().InitWithContext(tctx, &vapi.InitRequest{
		SecretShares:    cli.Flags.Init.UnsealKeys,
		SecretThreshold: cli.Flags.Init.UnsealThreshold,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize vault: %w", err)
	}

	config.RootToken = resp.RootToken
	config.UnsealKeys = resp.KeysB64
	SaveConfig()
	logger.InfoContext(ctx, "vault initialized")

	return nil
}

func WaitVaultUnseal(ctx context.Context, timeout time.Duration, node int) error {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	vault := NewVaultClient(ctx, node)

	for attempt := range 30 {
		l := logger.With(
			"node", node,
			"attempt", attempt,
		)

		if tctx.Err() != nil {
			return tctx.Err()
		}

		status, err := vault.Sys().SealStatusWithContext(tctx)
		if err != nil {
			l.ErrorContext(ctx, "failed to get health", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if !status.Sealed {
			l.InfoContext(ctx, "vault is already unsealed")
			break
		}

		if !status.Initialized {
			l.InfoContext(ctx, "vault is not initialized")
			time.Sleep(5 * time.Second)
			continue
		}

		l.InfoContext(ctx, "resetting unseal progress")
		_, err = vault.Sys().ResetUnsealProcessWithContext(tctx)
		if err != nil {
			l.ErrorContext(ctx, "failed to reset unseal process", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		issueUnsealing := false
		for keyIndex, key := range config.UnsealKeys {
			l.InfoContext(ctx, "using unseal key", "key_index", keyIndex)
			_, err = vault.Sys().UnsealWithContext(tctx, key)
			if err != nil {
				l.ErrorContext(ctx, "failed to use unseal key", "key_index", keyIndex, "error", err)
				issueUnsealing = true
				break
			}
		}

		if issueUnsealing {
			time.Sleep(5 * time.Second)
			continue
		}
	}

	return nil
}

func WaitVaultHealthy(ctx context.Context, timeout time.Duration, node int) error {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	vault := NewVaultClient(ctx, node)

	for attempt := range 30 {
		l := logger.With(
			"node", node,
			"attempt", attempt,
		)

		if tctx.Err() != nil {
			return tctx.Err()
		}

		health, err := vault.Sys().HealthWithContext(tctx)
		if err != nil {
			l.ErrorContext(ctx, "failed to get health", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if health.Sealed || !health.Initialized {
			l.ErrorContext(ctx, "vault is not healthy")
			time.Sleep(1 * time.Second)
			continue
		}

		l.InfoContext(ctx, "vault is healthy")
		break
	}

	return nil
}
