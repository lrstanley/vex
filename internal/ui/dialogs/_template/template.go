// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package template

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel

	// Core state.
	app types.AppState
	foo string

	// Styles.
	baseStyle lipgloss.Style
}

func New(app types.AppState) *Model {
	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeSmall,
			DisableChildren: true,
		},
		app: app,
	}

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.DialogFg())
}

func (m *Model) GetTitle() string {
	return "Example Template"
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width
	case styles.ThemeUpdatedMsg:
		m.initStyles()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel):
			return types.CloseActiveDialog()
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.baseStyle.Render(m.foo)
}
