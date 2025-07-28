// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package configstate

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/viewport"
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

	// Child components.
	code *viewport.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: false,
			RefreshInterval:  30 * time.Second,
		},
		app: app,
	}

	m.code = viewport.New(app)

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.code.Init(),
		types.RefreshData(m.UUID()),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case types.PageVisibleMsg:
		return types.RefreshData(m.UUID())
	case types.RefreshDataMsg:
		return tea.Batch(
			types.PageLoading(),
			m.app.Client().GetConfigState(m.UUID()),
		)
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		switch vmsg := msg.Msg.(type) {
		case types.ClientConfigStateMsg:
			cmds = append(cmds, types.PageLoaded())
			if msg.Error == nil {
				if vmsg.Data == nil {
					m.code.SetCode("No data available", "text")
				} else {
					m.code.SetJSON(vmsg.Data)
				}
			} else {
				m.code.SetError(msg.Error)
			}
		}
	}

	cmds = append(cmds, m.code.Update(msg))
	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return m.code.View()
}
