// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package statusbar

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/statusbar/filterelement"
	"github.com/lrstanley/vex/internal/ui/components/statusbar/statuselement"
	"github.com/lrstanley/vex/internal/ui/components/statusbar/vaultelement"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	Address     string
	isFiltering bool

	// Styles.
	baseStyle lipgloss.Style
	logoStyle lipgloss.Style

	// Child components.
	statusEl *statuselement.Model
	filterEl *filterelement.Model
	vaultEl  *vaultelement.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		statusEl:       statuselement.New(app),
		filterEl:       filterelement.New(app),
		vaultEl:        vaultelement.New(app),
	}

	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.BarFg()).
		Background(styles.Theme.BarBg())

	m.logoStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Background(styles.Theme.StatusBarLogoBg()).
		Foreground(styles.Theme.StatusBarLogoFg()).
		Bold(true)
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.statusEl.Init(),
		m.filterEl.Init(),
		m.vaultEl.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
		return tea.Sequence(
			m.statusEl.Update(msg),
			m.filterEl.Update(msg),
			m.vaultEl.Update(msg),
		)
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	case types.StatusMsg:
		return m.statusEl.Update(msg)
	case types.AppFocusChangedMsg:
		if msg.ID == types.FocusStatusBar {
			m.isFiltering = true
		} else {
			m.isFiltering = false
		}
	case tea.KeyMsg:
		if key.Matches(msg, types.KeyQuit) {
			return types.AppQuit()
		}
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		if m.isFiltering {
			return m.filterEl.Update(msg)
		}
		return nil
	}

	return tea.Batch(append(
		cmds,
		m.statusEl.Update(msg),
		m.filterEl.Update(msg),
		m.vaultEl.Update(msg),
	)...)
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}

	var status string

	if m.isFiltering {
		status = m.filterEl.View()
	} else {
		status = m.statusEl.View()
	}
	statusw := ansi.StringWidth(status)

	logo := m.logoStyle.Render("vex")
	logow := ansi.StringWidth(logo)

	m.vaultEl.Width = m.Width - logow
	vault := m.vaultEl.View()
	vaultw := ansi.StringWidth(vault)

	// This allows "overlapping" of the status on top of the regular statusbar elements.
	trailing := ansi.Cut(
		m.baseStyle.Render(strings.Repeat(" ", max(0, m.Width-vaultw-logow)))+vault+logo,
		statusw,
		m.Width,
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		status,
		trailing,
	)
}
