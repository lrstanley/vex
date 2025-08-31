// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package secrets

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/pages/kvv2versions"
	"github.com/lrstanley/vex/internal/ui/pages/kvviewsecret"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands = []string{"secrets", "secret"}
	columns  = []*table.Column{
		{ID: "full_path", Title: "Full Path"},
		{ID: "mount_type", Title: "Mount Type"},
		{ID: "capabilities", Title: "Capabilities"},
	}
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	filter string
	mount  *types.Mount // Optional mount to restrict to.
	data   *types.ClientListAllSecretsRecursiveMsg

	// Child components.
	table *table.Model[*table.StaticRow[*types.ClientSecretTreeRef]]
}

func New(app types.AppState, mount *types.Mount) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			RefreshInterval:  60 * time.Second,
		},
		app:   app,
		mount: mount,
	}

	m.table = table.New(app, columns, table.Config[*table.StaticRow[*types.ClientSecretTreeRef]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListAllSecretsRecursive(m.UUID(), m.mount)
		},
		SelectFn: func(value *table.StaticRow[*types.ClientSecretTreeRef]) tea.Cmd {
			return m.selectSecret(value.Value)
		},
		RowFn: func(value *table.StaticRow[*types.ClientSecretTreeRef]) []string {
			return []string{
				value.Value.GetFullPath(true),
				value.Value.Mount.Type,
				styles.ClientCapabilities(value.Value.Capabilities, value.Value.GetFullPath(true)),
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
		m.filter = msg.Text
		m.table.SetFilter(msg.Text)
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		if msg.Error != nil {
			return types.PageErrors(msg.Error)
		}

		switch vmsg := msg.Msg.(type) {
		case types.ClientListAllSecretsRecursiveMsg:
			cmds = append(cmds, types.PageClearState())

			m.data = &vmsg
			var data []*types.ClientSecretTreeRef

			for ref := range vmsg.Tree.IterRefs() {
				if !ref.IsSecret() {
					continue
				}
				data = append(data, ref)
			}

			slices.SortFunc(data, func(a, b *types.ClientSecretTreeRef) int {
				return strings.Compare(a.GetFullPath(true), b.GetFullPath(true))
			})

			m.table.SetRows(table.RowsFrom(data, func(d *types.ClientSecretTreeRef) table.ID {
				return table.ID(d.GetFullPath(true))
			}))
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) selectSecret(secret *types.ClientSecretTreeRef) tea.Cmd {
	if secret.Mount.KVVersion() == 2 {
		return types.OpenPage(kvv2versions.New(m.app, secret.Mount, secret.GetFullPath(false)), false)
	}
	return types.OpenPage(kvviewsecret.New(m.app, secret.Mount, secret.GetFullPath(false), 0), false)
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(m.table.TotalFilteredRows(), "secret", "secrets")
}

func (m *Model) TopRightBorder() string {
	if m.data == nil {
		return ""
	}
	out := fmt.Sprintf("requests: %d/%d", m.data.Requests, m.data.MaxRequests)

	if m.data.RequestAttempts >= m.data.MaxRequests {
		out += lipgloss.NewStyle().Foreground(styles.Theme.ErrorFg()).Render(" (max hit)")
	}

	return out
}

func (m *Model) GetTitle() string {
	if m.mount == nil {
		return "Secrets (recursive)"
	}

	return fmt.Sprintf("Secrets (recursive): %s", m.mount.Path)
}
