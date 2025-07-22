// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package statusbar

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/colors"
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/shorthelp"
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
	health      *vapi.HealthResponse
	isFiltering bool

	// Styles.
	baseStyle lipgloss.Style
	logoStyle lipgloss.Style

	// Child components.
	statusEl *statuselement.Model
	filterEl *filterelement.Model
	helpEl   *shorthelp.Model
	vaultEl  *vaultelement.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		statusEl:       statuselement.New(app),
		filterEl:       filterelement.New(app),
		helpEl:         shorthelp.New(app),
		vaultEl:        vaultelement.New(app),
	}

	m.setStyles()
	m.updateKeyBinds()
	return m
}

func (m *Model) setStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.StatusBarFg()).
		Background(styles.Theme.StatusBarBg())

	m.helpEl.Styles.Base = m.helpEl.Styles.Base.
		Background(styles.Theme.StatusBarBg()).
		Foreground(styles.Theme.StatusBarFg())
	m.helpEl.Styles.Key = m.helpEl.Styles.Key.
		Foreground(styles.Theme.ShortHelpKeyFg())
	m.helpEl.Styles.Desc = m.helpEl.Styles.Desc.
		Foreground(colors.Darken(styles.Theme.StatusBarFg(), 30))
	m.helpEl.Styles.Separator = m.helpEl.Styles.Separator.
		Foreground(colors.Darken(styles.Theme.StatusBarFg(), 30))

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
		m.helpEl.Init(),
		m.vaultEl.Init(),
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
	case types.StatusMsg:
		return m.statusEl.Update(msg)
	case types.AppFocusChangedMsg:
		if msg.ID == types.FocusStatusBar {
			m.isFiltering = true
		} else {
			m.isFiltering = false
		}
		m.updateKeyBinds()
	case tea.KeyMsg:
		if key.Matches(msg, types.KeyQuit) {
			return tea.Quit
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
		m.helpEl.Update(msg),
		m.vaultEl.Update(msg),
	)...)
}

func (m *Model) updateKeyBinds() {
	kb := []key.Binding{types.KeyCommander}

	if m.app.Page().Get().GetSupportFiltering() {
		kb = append(kb, types.KeyFilter)
	}

	kb = append(kb, types.KeyHelp)

	m.helpEl.SetKeyBinds(append(
		kb,
		m.app.Page().Get().ShortHelp()...,
	)...)
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}

	var out []string

	if m.isFiltering {
		out = append(out, m.filterEl.View())
	} else {
		out = append(out, m.statusEl.View())
	}

	vault := m.vaultEl.View()
	logo := m.logoStyle.Render("vex")

	available := m.Width - styles.W(append(out, vault, logo)...)

	help := m.helpEl.View()
	if styles.W(help) > available-2 {
		help = m.baseStyle.Width(available).Padding(0, 1).Render(styles.Trunc(help, available-2))
	} else {
		help = m.baseStyle.Width(available).Align(lipgloss.Right).Padding(0, 1).Render(help)
	}

	out = append(out, help)
	out = append(out, vault)
	out = append(out, logo)

	return lipgloss.JoinHorizontal(lipgloss.Left, out...)
}
