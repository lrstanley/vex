// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ui

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/tasks"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/statusbar"
	"github.com/lrstanley/vex/internal/ui/dialogs"
	"github.com/lrstanley/vex/internal/ui/dialogs/commander"
	"github.com/lrstanley/vex/internal/ui/dialogs/help"
	"github.com/lrstanley/vex/internal/ui/page"
	"github.com/lrstanley/vex/internal/ui/page/aclpolicies"
	"github.com/lrstanley/vex/internal/ui/page/configstate"
	"github.com/lrstanley/vex/internal/ui/page/mounts"
	"github.com/lrstanley/vex/internal/ui/page/secrets"
	"github.com/lrstanley/vex/internal/ui/styles"
)

func pageInitializer(app types.AppState) []commander.PageRef {
	return []commander.PageRef{
		{
			Description: "View mounts",
			Commands:    mounts.Commands,
			New: func() types.Page {
				return mounts.New(app)
			},
		},
		{
			Description: "View secrets",
			Commands:    secrets.Commands,
			New: func() types.Page {
				return secrets.New(app, nil, "")
			},
		},
		{
			Description: "View ACL policies",
			Commands:    aclpolicies.Commands,
			New: func() types.Page {
				return aclpolicies.New(app)
			},
		},
		{
			Description: "View config state",
			Commands:    configstate.Commands,
			New: func() types.Page {
				return configstate.New(app)
			},
		},
	}
}

type Model struct {
	// Core state, clients, etc.
	app types.AppState

	// UI state.
	height        int
	width         int
	focused       types.FocusID
	previousFocus types.FocusID
	cmdConfig     commander.Config

	// Sub-components.
	statusbar types.Component
}

func New(client types.Client) *Model {
	app := &appState{
		client: client,
		dialog: dialogs.NewState(),
		task:   tasks.NewState(),
	}

	app.page = page.NewState(mounts.New(app))

	return &Model{
		app: app,
		cmdConfig: commander.Config{
			App:   app,
			Pages: pageInitializer(app),
		},
		focused:   types.FocusPage,
		statusbar: statusbar.New(app),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Sequence(
		styles.Theme.Update(true),
		m.app.Page().Init(),
		tea.Batch(
			m.app.Client().Init(),
			m.app.Dialog().Init(),
			m.statusbar.Init(),
		),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("tui update", "message", fmt.Sprintf("%#v", msg), "type", fmt.Sprintf("%T", msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		return m, tea.Batch(
			m.app.Page().Update(tea.WindowSizeMsg{
				Height: msg.Height - 1, // -1=statusbar.
				Width:  msg.Width,
			}),
			m.app.Dialog().Update(msg),
			m.statusbar.Update(tea.WindowSizeMsg{
				Height: 1,
				Width:  msg.Width,
			}),
		)
	case tea.BackgroundColorMsg:
		cmds = append(cmds, styles.Theme.Update(msg.IsDark()))
	case tea.KeyMsg:
		switch {
		case msg.String() == "[":
			return m, styles.Theme.PreviousTint()
		case msg.String() == "]":
			return m, styles.Theme.NextTint()
		}

		switch m.focused {
		case types.FocusDialog:
			return m, m.app.Dialog().Update(msg)
		case types.FocusPage:
			if !m.app.Page().Get().HasInputFocus() {
				switch {
				case key.Matches(msg, types.KeyCommander):
					return m, types.OpenDialog(commander.New(m.app, m.cmdConfig))
				case key.Matches(msg, types.KeyFilter) && m.app.Page().Get().GetSupportFiltering():
					return m, types.FocusChange(types.FocusStatusBar)
				case key.Matches(msg, types.KeyHelp):
					return m, types.OpenDialog(help.New(m.app))
				}
			}

			return m, m.app.Page().Update(msg)
		case types.FocusStatusBar:
			return m, m.statusbar.Update(msg)
		}
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		switch m.focused {
		case types.FocusDialog:
			return m, m.app.Dialog().Update(msg)
		case types.FocusPage:
			return m, m.app.Page().Update(msg)
		case types.FocusStatusBar:
			return m, m.statusbar.Update(msg)
		}
	case types.AppFocusChangedMsg:
		m.previousFocus = m.focused
		m.focused = msg.ID
	case types.AppRequestPreviousFocusMsg:
		m.focused = m.previousFocus
		return m, types.FocusChange(m.focused)
	case types.StatusMsg:
		cmds = append(cmds, m.statusbar.Update(msg))
		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(append(
		cmds,
		m.app.Page().Update(msg),
		m.app.Dialog().Update(msg),
		m.app.Client().Update(msg),
		m.statusbar.Update(msg),
	)...)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	s := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				m.app.Page().View(),
				m.statusbar.View(),
			),
		)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(s),
	}

	if m.app.Dialog().Len() > 0 {
		layers = append(layers, m.app.Dialog().GetLayers()...)
	}

	return lipgloss.NewCanvas(layers...).Render()
}
