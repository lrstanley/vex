// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in the
// LICENSE file.

package api

import (
	"net/http"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

type raftConfigResponse struct {
	Config struct {
		Servers []*types.RaftConfigPeer `json:"servers"`
	} `json:"config"`
}

func (c *client) GetRaftConfig(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientRaftConfigMsg, error) {
		out, err := request[wrappedResponse[raftConfigResponse]](
			c,
			http.MethodGet,
			"/v1/sys/storage/raft/configuration",
			nil,
			nil,
		)
		if err != nil {
			return nil, err
		}

		slices.SortFunc(out.Data.Config.Servers, func(a, b *types.RaftConfigPeer) int {
			return strings.Compare(strings.ToLower(a.Address), strings.ToLower(b.Address))
		})
		return &types.ClientRaftConfigMsg{Peers: out.Data.Config.Servers}, nil
	})
}

func (c *client) RemoveRaftPeer(uuid, serverID string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientSuccessMsg, error) {
		_, err := c.api.Logical().Write("sys/storage/raft/remove-peer", map[string]any{
			"server_id": serverID,
		})
		if err != nil {
			return nil, err
		}
		return &types.ClientSuccessMsg{Message: "raft peer removed"}, nil
	})
}
