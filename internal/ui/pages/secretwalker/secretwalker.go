// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package secretwalker

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/dialogs/confirm"
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
			FullKeyBinds: [][]key.Binding{{
				types.OverrideHelp(types.KeyDetails, "view metadata (kv v2 only)"),
				types.KeyOpenEditor,
				types.KeyDelete,
			}},
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
			var pathValue string

			if strings.HasSuffix(row.Value.Path, "/") {
				pathValue = lipgloss.NewStyle().Bold(true).Foreground(styles.Theme.InfoFg()).Render(styles.IconFolder() + " " + row.Value.Path)
			} else {
				pathValue = styles.IconSecret() + " " + row.Value.Path
			}

			return []string{
				row.Value.Mount.Path,
				pathValue,
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
		switch {
		case key.Matches(msg, types.KeyDetails) && m.mount.KVVersion() == 2:
			if v, ok := m.table.GetSelectedRow(); ok && !strings.HasSuffix(v.Value.Path, "/") {
				return m.app.Client().GetKVv2Metadata(m.UUID(), v.Value.Mount, v.Value.Path)
			}
		case key.Matches(msg, types.KeyOpenEditor):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.editSecret(v.Value)
			}
		case key.Matches(msg, types.KeyDelete):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.deleteSecret(v.Value)
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
		return types.OpenPage(kvviewsecret.New(m.app, secret.Mount, secret.Path, 0, false), false)
	}
	return types.OpenPage(New(m.app, secret.Mount, secret.Path), false)
}

func (m *Model) editSecret(secret *types.SecretListRef) tea.Cmd {
	if strings.HasSuffix(secret.Path, "/") {
		return nil
	}
	return types.OpenPage(kvviewsecret.New(m.app, secret.Mount, secret.Path, 0, true), false)
}

func (m *Model) deleteSecret(secret *types.SecretListRef) tea.Cmd {
	if strings.HasSuffix(secret.Path, "/") {
		return nil
	}

	var confirmCmds []tea.Cmd
	if len(m.table.GetAllRows()) > 1 {
		confirmCmds = []tea.Cmd{
			m.app.Client().DeleteKVSecret(m.UUID(), secret.Mount, secret.Path),
			types.CloseActiveDialog(),
			types.RefreshData(m.UUID()),
		}
	} else {
		confirmCmds = []tea.Cmd{
			m.app.Client().DeleteKVSecret(m.UUID(), secret.Mount, secret.Path),
			types.CloseActiveDialog(),
			types.CloseActivePage(),
		}
	}

	return types.OpenDialog(confirm.New(m.app, confirm.Config{
		Title:         fmt.Sprintf("Delete %s", secret.FullPath()),
		Message:       "Are you sure you want to delete this secret? This cannot be undone.",
		AllowsBlur:    true,
		ConfirmStatus: types.Error,
		ConfirmFn: func() tea.Cmd {
			return tea.Sequence(confirmCmds...)
		},
		CancelFn: types.CloseActiveDialog,
	}))
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
