// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
)

type Client interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	GetHealth() tea.Cmd
	ListMounts(uuid string) tea.Cmd
	ListSecrets(uuid string, mount *Mount, path string) tea.Cmd
	GetConfigState(uuid string) tea.Cmd
}

// ClientMsg is a wrapper for any message relating to Vault API/client/etc responses.
type ClientMsg struct {
	UUID  string
	Msg   any
	Error error
}

// ClientRequestConfigMsg is a message to request the Vault configuration.
type ClientRequestConfigMsg struct{}

// ClientConfigMsg is a message containing the Vault configuration. Health field is
// nil if the last health check failed.
type ClientConfigMsg struct {
	Address string
	Health  *vapi.HealthResponse
}

type ClientListMountsMsg struct {
	Mounts []*Mount
}

type Mount struct {
	*vapi.MountOutput
	Path string
}

type ClientListSecretsMsg struct {
	Values []*SecretListRef
}

type SecretListRef struct {
	Mount *Mount
	Path  string
}

type ClientConfigStateMsg struct {
	Data json.RawMessage
}
