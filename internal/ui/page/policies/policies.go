// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package policies

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var Commands = []string{"policies", "policy"}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height int
	width  int
}

func New(app types.AppState) *Model {
	return &Model{
		PageModel: &types.PageModel{
			Commands:      Commands,
			ShortKeyBinds: []key.Binding{types.KeyQuit},
			FullKeyBinds:  [][]key.Binding{{types.KeyQuit}},
		},
		app: app,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if key.Matches(msg, types.KeyQuit) {
			return tea.Quit
		}
	}
	return nil
}

func (m *Model) View() string {
	s := lipgloss.NewStyle().
		Foreground(styles.Theme.Fg()).
		Background(styles.Theme.Bg()).
		Width(m.width).
		Height(m.height)

	return s.Render("POLICIES PAGE")
}
