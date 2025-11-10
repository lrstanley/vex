// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

func (c *client) GetHealth(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientConfigMsg, error) {
		health := c.health.Get()
		if health == nil {
			var err error
			health, err = c.api.Sys().Health()
			if err != nil {
				return nil, err
			}
			c.health.Set(health, HealthCheckInterval-(1*time.Second))
		}
		return &types.ClientConfigMsg{
			Address: c.api.Address(),
			Health:  health,
		}, nil
	})
}

func (c *client) GetConfigState(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientConfigStateMsg, error) {
		// No Go client method for this.
		data, err := request[json.RawMessage](
			c,
			http.MethodGet,
			"/v1/sys/config/state/sanitized",
			nil,
			nil,
		)
		if err != nil {
			return nil, err
		}
		return &types.ClientConfigStateMsg{Data: data}, nil
	})
}
