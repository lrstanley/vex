// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package template

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var ID types.ComponentID = "template"

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	foo string

	// Styles.
	baseStyle lipgloss.Style
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{ID: ID},
		app:            app,
	}

	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg()).
		Background(styles.Theme.Bg())
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}

	return m.foo
}
