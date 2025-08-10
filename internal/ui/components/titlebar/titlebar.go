// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package titlebar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/shorthelp"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const MaxTitleWidth = 30

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// Styles.
	baseStyle lipgloss.Style

	// Child components.
	help *shorthelp.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		help:           shorthelp.New(),
	}

	m.setStyles()
	m.updateKeyBinds()
	return m
}

func (m *Model) setStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.BarFg())

	helpStyles := shorthelp.Styles{}
	helpStyles.Base = helpStyles.Base.
		Foreground(styles.Theme.BarFg()).
		Padding(0, 1)
	helpStyles.Key = helpStyles.Key.
		Foreground(styles.Theme.ShortHelpKeyFg())
	helpStyles.Desc = helpStyles.Desc.
		Foreground(lipgloss.Darken(styles.Theme.BarFg(), 0.3))
	helpStyles.Separator = helpStyles.Separator.
		Foreground(lipgloss.Darken(styles.Theme.BarFg(), 0.3))
	m.help.SetStyles(helpStyles)
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.help.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	case types.AppFocusChangedMsg:
		m.updateKeyBinds()
	}

	return tea.Batch(append(
		cmds,
		m.help.Update(msg),
	)...)
}

func (m *Model) updateKeyBinds() {
	m.help.SetKeyBinds(m.app.Page().ShortHelp()...)
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}

	m.help.SetMaxWidth(m.Width)

	var title string
	titleText := formatter.Trunc(" "+m.app.Page().Get().GetTitle(), MaxTitleWidth)
	titlew := ansi.StringWidth(titleText)

	if hw := ansi.StringWidth(m.help.View()); m.Width-titlew-hw > 3 {
		title = styles.Title(
			titleText,
			m.Width-hw,
			styles.IconTitleGradient,
			styles.Theme.TitleFg(),
			styles.Theme.TitleFromFg(),
			styles.Theme.TitleToFg(),
		)
	} else {
		title = m.baseStyle.Foreground(styles.Theme.TitleFg()).Render(titleText)
	}

	m.help.SetMaxWidth(m.Width - ansi.StringWidth(title))
	help := m.baseStyle.Render(m.help.View())

	available := m.Width - ansi.StringWidth(title) - ansi.StringWidth(help)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		m.baseStyle.Render(strings.Repeat(" ", max(0, available))),
		help,
	)
}
