// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package secrets

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/datatable"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands    = []string{"secrets", "secret"}
	dataColumns = []string{"Mount", "Key"}
)

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
			Commands:         Commands,
			SupportFiltering: true,
			ShortKeyBinds:    []key.Binding{types.KeyCancel, types.KeyRefresh, types.KeyQuit},
			FullKeyBinds:     [][]key.Binding{{types.KeyCancel, types.KeyRefresh, types.KeyQuit}},
		},
		app:   app,
		mount: mount,
		path:  path,
	}

	m.table = datatable.New(app, datatable.Config[*types.SecretListRef]{
		FetchFn: func(app types.AppState) tea.Cmd {
			return app.Client().ListSecrets(m.UUID(), m.mount, m.path)
		},
		SelectFn: func(app types.AppState, value *types.SecretListRef) tea.Cmd {
			if !strings.HasSuffix(value.Path, "/") {
				return nil
			}
			return types.OpenPage(New(app, value.Mount, value.Path), false)
		},
		RowFn: func(app types.AppState, value *types.SecretListRef) []string {
			return []string{value.Mount.Path, value.Path}
		},
	})

	return m
}

func (m *Model) Init() tea.Cmd {
	return m.table.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel):
			switch {
			case m.filter != "":
				return types.ClearAppFilter()
			case m.app.Page().HasParent():
				return types.CloseActivePage()
			}
			return nil
		case key.Matches(msg, types.KeyQuit):
			return tea.Quit
		}
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
