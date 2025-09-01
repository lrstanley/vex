// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package secretwalker

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/dialogs/genericcode"
	"github.com/lrstanley/vex/internal/ui/pages/kvv2versions"
	"github.com/lrstanley/vex/internal/ui/pages/kvviewsecret"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var columns = []*table.Column{
	{ID: "mount", Title: "Mount"},
	{ID: "key", Title: "Key"},
	{ID: "capabilities", Title: "Capabilities"},
}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	mount *types.Mount
	path  string

	// Child components.
	table *table.Model[*table.StaticRow[*types.SecretListRef]]
}

func New(app types.AppState, mount *types.Mount, path string) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
			FullKeyBinds: [][]key.Binding{
				{types.OverrideHelp(types.KeyDetails, "view metadata (kv v2 only)")},
			},
		},
		app:   app,
		mount: mount,
		path:  path,
	}

	m.table = table.New(app, columns, table.Config[*table.StaticRow[*types.SecretListRef]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListSecrets(m.UUID(), m.mount, m.path)
		},
		SelectFn: func(value *table.StaticRow[*types.SecretListRef]) tea.Cmd {
			return m.selectSecret(value.Value)
		},
		RowFn: func(row *table.StaticRow[*types.SecretListRef]) []string {
			return []string{
				row.Value.Mount.Path,
				row.Value.Path,
				styles.ClientCapabilities(row.Value.Capabilities, row.Value.FullPath()),
			}
		},
	})

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.table.Init(),
		types.RefreshData(m.UUID()),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.table.Update(msg)
	case types.PageVisibleMsg:
		return types.RefreshData(m.UUID())
	case types.RefreshDataMsg:
		return tea.Batch(
			types.PageLoading(),
			m.table.Fetch(false),
		)
	case types.AppFilterMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		m.table.SetFilter(msg.Text)
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		if msg.Error != nil {
			return types.PageErrors(msg.Error)
		}

		switch vmsg := msg.Msg.(type) {
		case types.ClientListSecretsMsg:
			cmds = append(cmds, types.PageClearState())
			m.table.SetRows(table.RowsFrom(vmsg.Values, func(v *types.SecretListRef) table.ID {
				return table.ID(v.Mount.Path + v.Path)
			}))
		case types.ClientGetKVv2MetadataMsg:
			return types.OpenDialog(genericcode.NewYAML(
				m.app,
				"Metadata: "+vmsg.Mount.Path+vmsg.Path,
				false,
				vmsg.Metadata,
			))
		}
	case tea.KeyMsg:
		if key.Matches(msg, types.KeyDetails) && m.mount.KVVersion() == 2 {
			if v, ok := m.table.GetSelectedRow(); ok && !strings.HasSuffix(v.Value.Path, "/") {
				return m.app.Client().GetKVv2Metadata(m.UUID(), v.Value.Mount, v.Value.Path)
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) selectSecret(secret *types.SecretListRef) tea.Cmd {
	if !strings.HasSuffix(secret.Path, "/") {
		if secret.Mount.KVVersion() == 2 {
			return types.OpenPage(kvv2versions.New(m.app, secret.Mount, secret.Path), false)
		}
		return types.OpenPage(kvviewsecret.New(m.app, secret.Mount, secret.Path, 0), false)
	}
	return types.OpenPage(New(m.app, secret.Mount, secret.Path), false)
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func (m *Model) GetTitle() string {
	return m.mount.Path + m.path
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(m.table.TotalFilteredRows(), "secret", "secrets")
}
