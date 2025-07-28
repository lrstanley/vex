// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package template

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const ID types.DialogID = "template"

var _ types.Dialog[any] = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel[struct{}]

	// Core state.
	app types.AppState

	// UI state.
	foo string

	// Styles.
	baseStyle lipgloss.Style
}

func New(app types.AppState) *Model {
	m := &Model{
		DialogModel: &types.DialogModel[struct{}]{
			ID:              ID,
			Size:            types.DialogSizeSmall,
			DisableChildren: true,
			ShortKeyBinds:   []key.Binding{types.KeyPopNavigation, types.KeyQuit},
			FullKeyBinds:    [][]key.Binding{{types.KeyPopNavigation, types.KeyQuit}},
		},
		app: app,
	}

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Background(styles.Theme.Bg()).
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
		case key.Matches(msg, types.KeyCloseDialog):
			return types.CloseDialog(m)
		case key.Matches(msg, types.KeyQuit):
			return types.AppQuit()
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
