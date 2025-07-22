// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/sourcegraph/conc/pool"
)

func (c *client) ListSecrets(uuid string, mount *types.Mount, path string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListSecretsMsg, error) {
		var values []*types.SecretListRef

		if mount != nil {
			prefix := strings.TrimSuffix(mount.Path, "/")

			if ver, ok := mount.Options["version"]; ok && ver == "2" {
				prefix += "/metadata"
			}

			secret, err := c.api.Logical().List(prefix + "/" + path)
			if err != nil {
				return nil, fmt.Errorf("list secrets: %w", err)
			}

			for _, v := range secretToList(path, secret) {
				values = append(values, &types.SecretListRef{
					Mount: mount,
					Path:  v,
				})
			}

			return &types.ClientListSecretsMsg{Values: values}, nil
		}

		mountList, err := c.api.Sys().ListMounts()
		if err != nil {
			return nil, fmt.Errorf("list mounts: %w", err)
		}

		var mounts []*types.Mount
		for path, data := range mountList {
			mounts = append(mounts, &types.Mount{
				MountOutput: data,
				Path:        path,
			})
		}

		slices.SortFunc(mounts, func(a, b *types.Mount) int {
			return strings.Compare(a.Path, b.Path)
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		p := pool.NewWithResults[[]*types.SecretListRef]().
			WithContext(ctx).
			WithCancelOnError().
			WithMaxGoroutines(3)

		for _, mount := range mounts {
			p.Go(func(ctx context.Context) ([]*types.SecretListRef, error) {
				secret, err := c.api.Logical().List(mount.Path)
				if err != nil {
					return nil, err
				}

				var values []*types.SecretListRef
				for _, v := range secretToList(path, secret) {
					values = append(values, &types.SecretListRef{
						Mount: mount,
						Path:  v,
					})
				}
				return values, nil
			})
		}

		results, err := p.Wait()
		if err != nil {
			return nil, fmt.Errorf("list secrets: %w", err)
		}

		for _, value := range results {
			values = append(values, value...)
		}

		return &types.ClientListSecretsMsg{Values: values}, nil
	})
}
