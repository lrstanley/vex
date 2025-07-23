// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package secrets

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands    = []string{"secrets", "secret"}
	dataColumns = []string{"Mount", "Key"}
)

type Data struct {
	data *types.SecretListRef
}

func (d Data) Get() Data {
	return d
}

func (d Data) Row() []string {
	return []string{
		d.data.Mount.Path,
		d.data.Path,
	}
}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height         int
	width          int
	filter         string
	mount          *types.Mount
	path           string
	secrets        []*types.SecretListRef
	selectedSecret *types.SecretListRef

	// Child components.
	tableComponent *table.Model[Data]
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

	m.tableComponent = table.New(app, table.Config[Data]{
		OnSelect: func(item Data) {
			if !strings.HasSuffix(item.data.Path, "/") {
				return
			}
			m.selectedSecret = item.data
		},
	})

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.app.Client().ListSecrets(m.UUID(), m.mount, m.path),
		m.tableComponent.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		case key.Matches(msg, types.KeyRefresh):
			cmds = append(
				cmds,
				m.tableComponent.SetLoading(),
				m.app.Client().ListSecrets(m.UUID(), m.mount, m.path),
			)
		}
	case types.AppFilterMsg:
		if msg.UUID != m.UUID() {
			return nil
		}

		m.filter = msg.Text
		m.tableComponent.SetFilter(msg.Text)
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}

		switch vmsg := msg.Msg.(type) {
		case types.ClientListSecretsMsg:
			cmds = append(cmds, m.tableComponent.SetLoading())
			if msg.Error == nil {
				m.secrets = vmsg.Values
				m.updateTableData()
			}
		}
	}

	cmds = append(cmds, m.tableComponent.Update(msg))

	if v := m.selectedSecret; v != nil {
		m.selectedSecret = nil
		return types.OpenPage(New(m.app, v.Mount, v.Path), false)
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateTableData() {
	if len(m.secrets) == 0 {
		m.tableComponent.SetData(dataColumns, []Data{})
		return
	}

	var secretData []Data
	for _, v := range m.secrets {
		secretData = append(secretData, Data{data: v})
	}

	m.tableComponent.SetData(
		dataColumns,
		secretData,
	)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	return m.tableComponent.View()
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(len(m.secrets), "secret", "secrets")
}
