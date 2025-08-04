// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/viewport"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel

	// Core state.
	app types.AppState

	// Styles.
	titleStyle    lipgloss.Style
	keyStyle      lipgloss.Style
	keyInnerStyle lipgloss.Style
	descStyle     lipgloss.Style

	// Components.
	viewport *viewport.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeSmall,
			DisableChildren: true,
		},
		app:      app,
		viewport: viewport.New(app),
	}
	m.initStyles()
	m.generateHelp()
	return m
}

func (m *Model) initStyles() {
	m.titleStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg()).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Bold(true)

	m.keyStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg())

	m.keyInnerStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.ShortHelpKeyFg()).
		Bold(true)

	m.descStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg())
}

func (m *Model) GetTitle() string {
	return "Keybind Help"
}

func (m *Model) IsCoreDialog() bool {
	return true
}

func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width

		// If the viewport is smaller than the dialog height, resize the dialog
		// even smaller.
		m.Height = min(m.viewport.TotalLineCount(), m.Height)
		m.viewport.SetHeight(m.Height)
		m.viewport.SetWidth(m.Width)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		m.generateHelp()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyHelp):
			return types.CloseActiveDialog()
		}
	}

	return tea.Batch(append(
		cmds,
		m.viewport.Update(msg),
	)...)
}

func (m *Model) generateHelp() {
	var buf strings.Builder

	helpFocus := types.FocusPage
	if m.app.Dialog().Len() > 1 {
		helpFocus = types.FocusDialog
	}

	keys := m.app.FullHelp(helpFocus)

	var maxKeyWidth int
	for _, b := range keys {
		for _, binding := range b {
			maxKeyWidth = max(maxKeyWidth, len(binding.Help().Key))
		}
	}

	for _, bindings := range keys {
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}

		for _, binding := range bindings {
			buf.WriteString(
				m.keyStyle.Width(maxKeyWidth+4).Render(
					m.keyStyle.Render("<")+
						m.keyInnerStyle.Render(binding.Help().Key)+
						m.keyStyle.Render(">"),
				) +
					m.descStyle.Render(binding.Help().Desc) + "\n",
			)
		}
	}
	m.viewport.SetContent(strings.TrimSuffix(buf.String(), "\n"))
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.viewport.View()
}
