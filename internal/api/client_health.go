// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

func (c *client) GetHealth() tea.Cmd {
	return func() tea.Msg {
		health := c.health.Get()
		if health == nil {
			var err error
			health, err = c.api.Sys().Health()
			if err != nil {
				return types.ClientMsg{
					Msg:   types.ClientConfigMsg{Address: c.api.Address()},
					Error: fmt.Errorf("get health: %w", err),
				}
			}
			c.health.Set(health, HealthCheckInterval-(1*time.Second))
		}
		return types.ClientMsg{Msg: types.ClientConfigMsg{
			Address: c.api.Address(),
			Health:  health,
		}}
	}
}

func (c *client) GetConfigState(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientConfigStateMsg, error) {
		// No Go client method for this.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(c.api.Address(), "/")+"/v1/sys/config/state/sanitized", http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("get config state: %w", err)
		}

		req.Header = c.api.Headers()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("get config state: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("get config state: %s", resp.Status)
		}

		var data json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("get config state: %w", err)
		}

		return &types.ClientConfigStateMsg{Data: data}, nil
	})
}
