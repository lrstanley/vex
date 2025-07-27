// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"golang.org/x/sync/errgroup"
)

const MaxRecursiveRequests = 200

func (c *client) ListSecrets(uuid string, mount *types.Mount, path string) tea.Cmd {
	// TODO: refactor and dedup logic.
	return wrapHandler(uuid, func() (*types.ClientListSecretsMsg, error) {
		var values []*types.SecretListRef

		paths, err := c.list(mount, path)
		if err != nil {
			return nil, fmt.Errorf("list secrets: %w", err)
		}

		for _, v := range paths {
			values = append(values, &types.SecretListRef{
				Mount: mount,
				Path:  path + v,
			})
		}

		return &types.ClientListSecretsMsg{Values: values}, nil
	})
}

// list lists the keys under a given path. Make sure to normalize the path, to not
// include the mount path, "metadata/" for KVv2, etc.
func (c *client) list(mount *types.Mount, path string) (values []string, err error) {
	prefix := strings.TrimSuffix(mount.Path, "/")

	if mount.KVVersion() == 2 {
		prefix += "/metadata"
	}

	secret, err := c.api.Logical().List(prefix + "/" + path)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	return secretToList(secret), nil
}

func (c *client) listAllSecretsRecursive(maxRequests int64, withCapabilities bool) (tree types.ClientSecretTree, requests int64, err error) {
	mounts, err := c.listMounts(true)
	if err != nil {
		return nil, 0, err
	}

	var mu sync.Mutex
	var eg errgroup.Group
	var req atomic.Int64

	for _, mount := range mounts {
		if mount.Type != "kv" {
			continue
		}

		eg.Go(func() error {
			inner, err := c.listMountSecretsRecursive(&req, maxRequests, withCapabilities, mount, "")
			if err != nil {
				return err
			}
			mu.Lock()
			tree = append(tree, inner...)
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, req.Load(), err
	}

	slices.SortFunc(tree, func(a, b *types.ClientSecretTreeRef) int {
		return strings.Compare(a.Path, b.Path)
	})

	return tree, req.Load(), nil
}

// listMountSecretsRecursive lists the secrets for a given mount.
func (c *client) listMountSecretsRecursive(
	req *atomic.Int64,
	maxRequests int64,
	withCapabilities bool,
	mount *types.Mount,
	parent string,
	paths ...string,
) (tree types.ClientSecretTree, err error) {
	var mu sync.Mutex
	var eg errgroup.Group

	wasMountLevel := false

	if len(paths) == 0 {
		wasMountLevel = true

		if v := req.Add(1); v >= maxRequests {
			tree = append(tree, &types.ClientSecretTreeRef{
				Mount:      mount,
				Path:       mount.Path,
				Incomplete: true,
			})
			return tree, nil
		}

		parent = ""
		paths, err = c.list(mount, "")
		if err != nil {
			return nil, err
		}
	}

	for _, path := range paths {
		switch {
		case strings.HasSuffix(path, "/"): // Folder.
			eg.Go(func() error {
				if v := req.Add(1); v >= maxRequests {
					ref := &types.ClientSecretTreeRef{
						Mount:      mount,
						Path:       path,
						Incomplete: true,
					}

					mu.Lock()
					tree = append(tree, ref)
					mu.Unlock()
					return nil
				}

				ipaths, err := c.list(mount, parent+path)
				if err != nil {
					return err
				}

				if len(ipaths) == 0 {
					return nil
				}

				inner, err := c.listMountSecretsRecursive(req, maxRequests, false, mount, parent+path, ipaths...)
				if err != nil {
					return err
				}

				ref := &types.ClientSecretTreeRef{
					Mount: mount,
					Path:  path,
					Leafs: inner,
				}

				for _, leaf := range inner {
					leaf.Parent = ref

					if leaf.Incomplete {
						ref.Incomplete = true
					}
				}

				mu.Lock()
				tree = append(tree, ref)
				mu.Unlock()

				return nil
			})
		case !strings.HasSuffix(path, "/"): // Secret.
			mu.Lock()
			ref := &types.ClientSecretTreeRef{
				Mount: mount,
				Path:  path,
			}

			tree = append(tree, ref)
			mu.Unlock()
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	slices.SortFunc(tree, func(a, b *types.ClientSecretTreeRef) int {
		return strings.Compare(a.Path, b.Path)
	})

	if wasMountLevel {
		tree = []*types.ClientSecretTreeRef{{
			Mount: mount,
			Path:  mount.Path,
			Leafs: tree,
		}}

		for i := range tree[0].Leafs {
			tree[0].Leafs[i].Parent = tree[0]

			if tree[0].Leafs[i].Incomplete {
				tree[0].Incomplete = true
			}
		}
	}

	// Rather than fetching capabilities for each path individually, we recursively
	// fetch everything we need, then query for the capabilities in 1 go. The
	// downside is that we need to re-apply the capabilities against the tree
	// structure, which feels like a bit of a hack, however, this is way more
	// efficient in terms of HTTP requests. We also don't count this request
	// towards the maxRequests limit.
	if withCapabilities {
		var values []string
		for ref := range tree.IterRefs() {
			values = append(values, ref.GetFullPath())
		}

		capabilities, err := c.getCapabilities(values...)
		if err != nil {
			return nil, err
		}

		for path, caps := range capabilities {
			for ref := range tree.IterRefs() {
				if ref.ApplyCapabilities(path, caps) {
					break
				}
			}
		}
	}

	return tree, nil
}

func (c *client) ListAllSecretsRecursive(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListAllSecretsRecursiveMsg, error) {
		tree, requests, err := c.listAllSecretsRecursive(MaxRecursiveRequests, true)
		if err != nil {
			return nil, err
		}
		return &types.ClientListAllSecretsRecursiveMsg{
			Tree:        tree,
			Requests:    requests,
			MaxRequests: MaxRecursiveRequests,
		}, nil
	})
}

func (c *client) ListMountSecretsRecursive(uuid string, mount *types.Mount, path string) tea.Cmd {
	return nil
}
