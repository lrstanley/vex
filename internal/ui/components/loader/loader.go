// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package loader

import (
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Styles.
	baseStyle lipgloss.Style

	// Child components.
	spinner spinner.Model
}

func New() *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		spinner:        spinner.New(),
	}

	m.spinner.Spinner = spinner.MiniDot

	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg()).
		Padding(0, 1).
		Align(lipgloss.Center)

	m.spinner.Style = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg())
}

func (m *Model) Init() tea.Cmd {
	return m.Active()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) SetHeight(height int) {
	m.Height = height
}

func (m *Model) SetWidth(width int) {
	m.Width = width
}

func (m *Model) Active() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}
	return m.baseStyle.
		Width(m.Width).
		Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(m.spinner.View() + m.baseStyle.Render("loading..."))
}
