// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/types"
	"golang.org/x/sync/errgroup"
)

const MaxRecursiveRequests = 300

func (c *client) ListSecrets(uuid string, mount *types.Mount, path string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListSecretsMsg, error) {
		var values []*types.SecretListRef

		paths, err := c.list(mount, path)
		if err != nil {
			return nil, fmt.Errorf("list secrets: %w", err)
		}

		capabilities, err := c.getCapabilities(mount.PrefixPaths(paths...)...)
		if err != nil {
			return nil, fmt.Errorf("get capabilities: %w", err)
		}

		for _, v := range paths {
			values = append(values, &types.SecretListRef{
				Mount:        mount,
				Path:         path + v,
				Capabilities: capabilities[mount.Path+v],
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

func (c *client) listAllSecretsRecursive(
	mount *types.Mount,
	maxRequests int64,
	withCapabilities bool,
) (tree types.ClientSecretTree, requestAttempts, requests int64, err error) {
	var mounts []*types.Mount

	if mount != nil {
		mounts = append(mounts, mount)
	} else {
		mounts, err = c.listMounts(true)
	}

	if err != nil {
		return nil, 0, 0, err
	}

	var mu sync.Mutex
	var eg errgroup.Group
	var reqAttempts, actualRequests atomic.Int64

	for _, mount := range mounts {
		if mount.Type != "kv" {
			continue
		}

		eg.Go(func() error {
			inner, eerr := c.listMountSecretsRecursive(
				&reqAttempts,
				&actualRequests,
				maxRequests,
				withCapabilities,
				mount,
				"",
			)
			if eerr != nil {
				return eerr
			}
			mu.Lock()
			tree = append(tree, inner...)
			mu.Unlock()
			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return nil, reqAttempts.Load(), actualRequests.Load(), err
	}

	slices.SortFunc(tree, func(a, b *types.ClientSecretTreeRef) int {
		return strings.Compare(a.Path, b.Path)
	})

	return tree, reqAttempts.Load(), actualRequests.Load(), nil
}

// listMountSecretsRecursive lists the secrets for a given mount.
func (c *client) listMountSecretsRecursive( //nolint:gocognit
	reqAttempts *atomic.Int64,
	actualRequests *atomic.Int64,
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

		if v := reqAttempts.Add(1); v > maxRequests {
			tree = append(tree, &types.ClientSecretTreeRef{
				Mount:      mount,
				Path:       mount.Path,
				Incomplete: true,
			})
			return tree, nil
		}

		parent = ""
		actualRequests.Add(1)
		paths, err = c.list(mount, "")
		if err != nil {
			return nil, err
		}
	}

	for _, path := range paths {
		switch {
		case strings.HasSuffix(path, "/"): // Folder.
			eg.Go(func() error {
				if v := reqAttempts.Add(1); v > maxRequests {
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

				var ipaths []string
				actualRequests.Add(1)
				ipaths, err = c.list(mount, parent+path)
				if err != nil {
					return err
				}

				if len(ipaths) == 0 {
					return nil
				}

				var inner types.ClientSecretTree
				inner, err = c.listMountSecretsRecursive(
					reqAttempts,
					actualRequests,
					maxRequests,
					false,
					mount,
					parent+path,
					ipaths...,
				)
				if err != nil {
					return err
				}

				ref := &types.ClientSecretTreeRef{
					Mount: mount,
					Path:  path,
					Leafs: inner,
				}

				ref.Leafs.SetParentOnLeafs(ref)

				for _, leaf := range inner {
					if leaf.Incomplete {
						ref.Incomplete = true
						break
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

	err = eg.Wait()
	if err != nil {
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

		tree[0].Leafs.SetParentOnLeafs(tree[0])

		for i := range tree[0].Leafs {
			if tree[0].Leafs[i].Incomplete {
				tree[0].Incomplete = true
				break
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
			values = append(values, ref.GetFullPath(true))
		}

		var capabilities map[string]types.ClientCapabilities
		capabilities, err = c.getCapabilities(values...)
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

func (c *client) ListAllSecretsRecursive(uuid string, mount *types.Mount) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListAllSecretsRecursiveMsg, error) {
		tree, requestAttempts, requests, err := c.listAllSecretsRecursive(
			mount,
			MaxRecursiveRequests,
			true,
		)
		if err != nil {
			return nil, err
		}
		tree.SetParentOnLeafs(nil)
		return &types.ClientListAllSecretsRecursiveMsg{
			Tree:            tree,
			RequestAttempts: requestAttempts,
			Requests:        requests,
			MaxRequests:     MaxRecursiveRequests,
		}, nil
	})
}

func (c *client) GetKVv2Metadata(uuid string, mount *types.Mount, path string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientGetKVv2MetadataMsg, error) {
		if mount.KVVersion() != 2 {
			return nil, fmt.Errorf("get secret metadata: %w", errors.New("mount is not a kv v2 mount"))
		}
		metadata, err := c.api.KVv2(mount.Path).GetMetadata(context.Background(), path)
		if err != nil {
			return nil, fmt.Errorf("get secret metadata: %w", err)
		}
		return &types.ClientGetKVv2MetadataMsg{
			Mount:    mount,
			Path:     path,
			Metadata: metadata,
		}, nil
	})
}

func (c *client) ListKVv2Versions(uuid string, mount *types.Mount, path string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListKVv2VersionsMsg, error) {
		if mount.KVVersion() != 2 {
			return nil, fmt.Errorf("list kv v2 versions: %w", errors.New("mount is not a kv v2 mount"))
		}
		versions, err := c.api.KVv2(mount.Path).GetVersionsAsList(context.Background(), path)
		if err != nil {
			return nil, fmt.Errorf("list kv v2 versions: %w", err)
		}
		return &types.ClientListKVv2VersionsMsg{
			Mount:    mount,
			Path:     path,
			Versions: versions,
		}, nil
	})
}

func (c *client) GetKVSecret(uuid string, mount *types.Mount, path string, version int) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientGetSecretMsg, error) {
		var secret *vapi.KVSecret
		var err error
		if mount.KVVersion() == 2 {
			if version < 1 {
				secret, err = c.api.KVv2(mount.Path).Get(context.Background(), path)
			} else {
				secret, err = c.api.KVv2(mount.Path).GetVersion(context.Background(), path, version)
			}
		} else {
			secret, err = c.api.KVv1(mount.Path).Get(context.Background(), path)
		}

		if err != nil {
			return nil, fmt.Errorf("get secret: %w", err)
		}

		var data map[string]any
		if secret != nil && secret.Data != nil {
			data = secret.Data

			if v, ok := data["data"]; ok && mount.KVVersion() == 2 {
				if vv, vok := v.(map[string]any); vok {
					data = vv
				}
			}
		}

		return &types.ClientGetSecretMsg{
			Mount: mount,
			Path:  path,
			Data:  data,
		}, nil
	})
}

func (c *client) PutKVSecret(uuid string, mount *types.Mount, path string, data map[string]any) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientSuccessMsg, error) {
		var err error
		if mount.KVVersion() == 1 {
			err = c.api.KVv1(mount.Path).Put(context.Background(), path, data)
		} else {
			_, err = c.api.KVv2(mount.Path).Put(context.Background(), path, data, vapi.WithMergeMethod("rw"))
		}
		if err != nil {
			return nil, fmt.Errorf("put secret: %w", err)
		}
		return &types.ClientSuccessMsg{Message: "updated secret"}, nil
	})
}

func (c *client) DeleteKVSecret(uuid string, mount *types.Mount, path string, versions ...int) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientSuccessMsg, error) {
		var err error
		if mount.KVVersion() != 2 {
			err = c.api.KVv1(mount.Path).Delete(context.Background(), path)
		} else if len(versions) == 0 {
			err = c.api.KVv2(mount.Path).Delete(context.Background(), path)
		} else {
			err = c.api.KVv2(mount.Path).DeleteVersions(context.Background(), path, versions)
		}
		if err != nil {
			return nil, fmt.Errorf("delete secret: %w", err)
		}
		return &types.ClientSuccessMsg{Message: "deleted secret"}, nil
	})
}

func (c *client) UndeleteKVSecret(uuid string, mount *types.Mount, path string, versions ...int) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientSuccessMsg, error) {
		if mount.KVVersion() != 2 {
			return nil, fmt.Errorf("undelete secret: %w", errors.New("not a kv v2 mount"))
		}

		if len(versions) == 0 {
			results, err := c.api.KVv2(mount.Path).GetVersionsAsList(context.Background(), path)
			if err != nil {
				return nil, fmt.Errorf("undelete secret: %w", err)
			}
			if len(results) == 0 {
				return nil, errors.New("no versions to undelete")
			}
			versions = append(versions, results[0].Version)
		}

		err := c.api.KVv2(mount.Path).Undelete(context.Background(), path, versions)
		if err != nil {
			return nil, fmt.Errorf("undelete secret: %w", err)
		}
		return &types.ClientSuccessMsg{Message: "undeleted secret"}, nil
	})
}

func (c *client) DestroyKVSecret(uuid string, mount *types.Mount, path string, versions ...int) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientSuccessMsg, error) {
		if mount.KVVersion() != 2 {
			err := c.api.KVv1(mount.Path).Delete(context.Background(), path)
			if err != nil {
				return nil, fmt.Errorf("destroy secret: %w", err)
			}
			return &types.ClientSuccessMsg{Message: "destroyed secret"}, nil
		}

		var err error
		if len(versions) == 0 {
			err = c.api.KVv2(mount.Path).DeleteMetadata(context.Background(), path)
		} else {
			err = c.api.KVv2(mount.Path).Destroy(context.Background(), path, versions)
		}
		if err != nil {
			return nil, fmt.Errorf("destroy secret: %w", err)
		}
		return &types.ClientSuccessMsg{Message: "destroyed secret"}, nil
	})
}
