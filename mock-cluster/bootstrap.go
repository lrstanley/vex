// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"os"
	"strings"

	"github.com/go-faker/faker/v4"
	vapi "github.com/hashicorp/vault/api"
)

func BootstrapVaultAuthEngines(ctx context.Context) {
	vault := NewVaultClient(ctx, 1)

	mounts, err := vault.Sys().ListAuthWithContext(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list auth mounts", "error", err)
		os.Exit(1)
	}

	if _, ok := mounts["userpass/"]; !ok {
		logger.InfoContext(ctx, "enabling userpass engine")
		err = vault.Sys().EnableAuthWithOptionsWithContext(
			ctx,
			"userpass",
			&vapi.EnableAuthOptions{
				Type: "userpass",
				Config: vapi.MountConfigInput{
					ListingVisibility: "unauth",
					TokenType:         "batch",
				},
			},
		)
		if err != nil {
			logger.ErrorContext(ctx, "failed to enable userpass engine", "error", err)
			os.Exit(1)
		}
	}

	if _, ok := mounts["approle/"]; !ok {
		logger.InfoContext(ctx, "enabling approle engine")
		err = vault.Sys().EnableAuthWithOptionsWithContext(
			ctx,
			"approle",
			&vapi.EnableAuthOptions{
				Type: "approle",
			},
		)
		if err != nil {
			logger.ErrorContext(ctx, "failed to enable approle engine", "error", err)
			os.Exit(1)
		}
	}
}

func BootstrapVaultPolicy(ctx context.Context, name, policy string) {
	vault := NewVaultClient(ctx, 1)

	logger.InfoContext(ctx, "creating policy", "name", name)
	err := vault.Sys().PutPolicyWithContext(ctx, name, strings.TrimSpace(policy))
	if err != nil {
		logger.ErrorContext(ctx, "failed to create policy", "name", name, "error", err)
		os.Exit(1)
	}
}

func BootstrapVaultUser(ctx context.Context, username string, policies ...string) {
	vault := NewVaultClient(ctx, 1)

	pw := faker.Password()

	logger.InfoContext(ctx, "creating user", "username", username)
	_, err := vault.Logical().Write("auth/userpass/users/"+username, map[string]any{
		"password": pw,
		"policies": strings.Join(policies, ","),
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to create user", "username", username, "error", err)
		os.Exit(1)
	}

	config.SetUser(username, pw)
}

func BootstrapVaultMount(ctx context.Context, path string, config vapi.MountInput) {
	vault := NewVaultClient(ctx, 1)

	mounts, err := vault.Sys().ListMountsWithContext(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list mounts", "error", err)
		os.Exit(1)
	}

	if _, ok := mounts[path+"/"]; ok {
		logger.InfoContext(ctx, "mount already exists", "path", path)
		return
	}

	logger.InfoContext(ctx, "enabling mount", "path", path)
	err = vault.Sys().MountWithContext(ctx, path, &config)
	if err != nil {
		logger.ErrorContext(ctx, "failed to enable mount", "path", path, "error", err)
		os.Exit(1)
	}
}

func BootstrapVaultKVSecrets(ctx context.Context, path string, data map[string]any) {
	vault := NewVaultClient(ctx, 1)

	logger.InfoContext(ctx, "writing kv secrets", "path", path)
	_, err := vault.Logical().Write(path, data)
	if err != nil {
		logger.ErrorContext(ctx, "failed to write kv secrets", "path", path, "error", err)
		os.Exit(1)
	}
}
