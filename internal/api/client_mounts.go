// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

func (c *client) ListMounts(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListMountsMsg, error) {
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

		return &types.ClientListMountsMsg{Mounts: mounts}, nil
	})
}
