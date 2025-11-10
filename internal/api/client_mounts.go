// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/types"
)

func (c *client) ListMounts(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListMountsMsg, error) {
		mounts, err := c.listMounts(true)
		if err != nil {
			return nil, err
		}

		if len(mounts) > 0 {
			paths := make([]string, 0, len(mounts))
			for _, mount := range mounts {
				paths = append(paths, mount.Path)
			}

			var capabilities map[string]types.ClientCapabilities
			capabilities, err = c.getCapabilities(paths...)
			if err != nil {
				return nil, fmt.Errorf("get capabilities: %w", err)
			}

			for _, mount := range mounts {
				mount.Capabilities = capabilities[mount.Path]
			}
		}

		return &types.ClientListMountsMsg{Mounts: mounts}, nil
	})
}

func (c *client) listMounts(ui bool) (mounts []*types.Mount, err error) {
	var mountList map[string]*api.MountOutput

	if ui {
		var data *wrappedResponse[map[string]map[string]*api.MountOutput]
		data, err = request[*wrappedResponse[map[string]map[string]*api.MountOutput]](
			c,
			http.MethodGet,
			"/v1/sys/internal/ui/mounts",
			nil,
			nil,
		)
		if err != nil {
			return nil, err
		}

		mountList = make(map[string]*api.MountOutput)

		for _, mounts := range data.Data {
			maps.Copy(mountList, mounts)
		}
	} else {
		mountList, err = c.api.Sys().ListMounts()
		if err != nil {
			return nil, fmt.Errorf("list mounts: %w", err)
		}
	}

	for path, data := range mountList {
		mounts = append(mounts, &types.Mount{
			MountOutput: data,
			Path:        path,
		})
	}

	slices.SortFunc(mounts, func(a, b *types.Mount) int {
		return strings.Compare(a.Path, b.Path)
	})

	return mounts, nil
}
