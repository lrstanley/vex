// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package secretwalker

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/datatable"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var dataColumns = []string{"Mount", "Key"}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	filter string
	mount  *types.Mount
	path   string

	// Child components.
	table *datatable.Model[*types.SecretListRef]
}

func New(app types.AppState, mount *types.Mount, path string) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
		},
		app:   app,
		mount: mount,
		path:  path,
	}

	m.table = datatable.New(app, datatable.Config[*types.SecretListRef]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListSecrets(m.UUID(), m.mount, m.path)
		},
		SelectFn: func(value *types.SecretListRef) tea.Cmd {
			if !strings.HasSuffix(value.Path, "/") {
				return nil
			}
			return types.OpenPage(New(app, value.Mount, value.Path), false)
		},
		RowFn: func(value *types.SecretListRef) []string {
			return []string{value.Mount.Path, value.Path}
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
		return m.table.Fetch(true)
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
		switch vmsg := msg.Msg.(type) {
		case types.ClientListSecretsMsg:
			cmds = append(cmds, m.table.SetLoading())
			if msg.Error == nil {
				m.table.SetData(
					dataColumns,
					vmsg.Values,
				)
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(m.table.DataLen(), "secret", "secrets")
}
