// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package configstate

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/code"
)

var Commands = []string{"configstate"}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height int
	width  int
	data   json.RawMessage

	// Child components.
	code *code.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: false,
			ShortKeyBinds:    []key.Binding{types.KeyCancel, types.KeyRefresh, types.KeyQuit},
			FullKeyBinds:     [][]key.Binding{{types.KeyCancel, types.KeyRefresh, types.KeyQuit}},
		},
		app: app,
	}

	m.code = code.New(app)

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.app.Client().GetConfigState(m.UUID()),
		m.code.Init(),
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
			if m.app.Page().HasParent() {
				return types.CloseActivePage()
			}
			return nil
		case key.Matches(msg, types.KeyQuit):
			return tea.Quit
		case key.Matches(msg, types.KeyRefresh):
			return m.app.Client().GetConfigState(m.UUID())
		}
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}

		switch vmsg := msg.Msg.(type) {
		case types.ClientConfigStateMsg:
			if msg.Error == nil {
				m.data = vmsg.Data
				m.updateCodeComponent()
			} else {
				m.code.SetError(msg.Error)
			}
		}
	}

	cmds = append(cmds, m.code.Update(msg))
	return tea.Batch(cmds...)
}

func (m *Model) updateCodeComponent() {
	if m.data == nil {
		m.code.SetCode("No data available", "text")
		return
	}
	m.code.SetJSON(m.data)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return m.code.View()
}
