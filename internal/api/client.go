// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/logging"
	"github.com/lrstanley/vex/internal/types"
)

const (
	HealthCheckInterval = 5 * time.Second
	TokenLookupInterval = 1 * time.Minute
)

var _ types.Client = &client{} // Ensure client implements types.Client.

type client struct {
	api  *vapi.Client
	http *http.Client

	firstHealthChecked atomic.Bool
	health             types.AtomicExpires[vapi.HealthResponse]
}

func NewClient() (types.Client, error) {
	c := &client{}

	cfg := vapi.DefaultConfig()
	cfg.MaxRetries = 5
	cfg.DisableRedirects = false
	cfg.Timeout = 5 * time.Second
	cfg.HttpClient.Transport = NewConcurrentLimiter(10, &logging.HTTPRoundTripper{
		RoundTripper: cfg.HttpClient.Transport,
	})

	if cfg.Error != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", cfg.Error)
	}

	c.http = cfg.HttpClient
	c.http.CheckRedirect = nil

	vc, err := vapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	c.api = vc
	return c, nil
}

func (c *client) Init() tea.Cmd {
	return tea.Batch(
		c.GetHealth(""),
		c.TokenLookupSelf(""),
	)
}

func (c *client) Update(msg tea.Msg) tea.Cmd {
	vm, ok := msg.(types.ClientMsg)
	if !ok {
		return nil
	}

	var cmds []tea.Cmd

	if vm.Error != nil {
		cmds = append(cmds, types.SendStatus(vm.Error.Error(), types.Error, 2*time.Second))
	} else {
		cmds = append(cmds, types.SendStatus("req success", types.Success, 250*time.Millisecond))
	}

	switch msg := vm.Msg.(type) {
	case types.ClientRequestConfigMsg:
		return c.GetHealth("")
	case types.ClientConfigMsg:
		if vm.UUID == "" {
			cmds = append(cmds, types.CmdAfterDuration(c.GetHealth(""), HealthCheckInterval))
		}

		if !c.firstHealthChecked.Load() && msg.Health != nil {
			c.firstHealthChecked.Store(true)
			cmds = append(cmds, types.SendStatus("initial health checked", types.Success, 1*time.Second))
		}
	case types.ClientTokenLookupSelfMsg:
		if vm.UUID == "" {
			cmds = append(cmds, types.CmdAfterDuration(c.TokenLookupSelf(""), TokenLookupInterval))
		}
	}

	return tea.Batch(cmds...)
}
