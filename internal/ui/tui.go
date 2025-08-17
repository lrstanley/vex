// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ui

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/config"
	"github.com/lrstanley/vex/internal/debouncer"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/statusbar"
	"github.com/lrstanley/vex/internal/ui/components/titlebar"
	"github.com/lrstanley/vex/internal/ui/dialogs/commander"
	"github.com/lrstanley/vex/internal/ui/dialogs/help"
	"github.com/lrstanley/vex/internal/ui/pages/aclpolicies"
	"github.com/lrstanley/vex/internal/ui/pages/configstate"
	"github.com/lrstanley/vex/internal/ui/pages/mounts"
	"github.com/lrstanley/vex/internal/ui/pages/secrets"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/vex/internal/ui/styles"
)

// Absolute minimum window size. If below this size, we display a message.
const (
	MinWinHeight = 13
	MinWinWidth  = 45
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
			Description: "View secrets (recursively)",
			Commands:    secrets.Commands,
			New: func() types.Page {
				return secrets.New(app)
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

var lastMouseEvent time.Time

func DownsampleMouseEvents(_ tea.Model, msg tea.Msg) tea.Msg {
	switch msg := msg.(type) {
	case tea.MouseWheelMsg, tea.MouseMotionMsg:
	case tea.KeyPressMsg:
		if msg.String() != "up" && msg.String() != "down" {
			return msg
		}
	default:
		return msg
	}

	now := time.Now()
	if now.Sub(lastMouseEvent) < 15*time.Millisecond {
		return nil
	}
	lastMouseEvent = now
	return msg
}

type Model struct { //nolint:recvcheck
	// Core state, clients, etc.
	app       types.AppState
	debouncer *debouncer.Service

	// UI state.
	height        int
	width         int
	focused       types.FocusID
	previousFocus types.FocusID
	cmdConfig     commander.Config

	// Sub-components.
	titlebar  types.Component
	statusbar types.Component
}

func New(client types.Client) *Model {
	app := &state.AppState{}
	app.SetClient(client)
	app.SetDialog(state.NewDialogState())
	app.SetPage(state.NewPageState(mounts.New(app)))

	return &Model{
		app:       app,
		debouncer: debouncer.New(),
		cmdConfig: commander.Config{
			App:   app,
			Pages: pageInitializer(app),
		},
		focused:   types.FocusPage,
		titlebar:  titlebar.New(app),
		statusbar: statusbar.New(app),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Sequence(
		m.debouncer.Init(),
		styles.Theme.Init(),
		m.app.Page().Init(),
		tea.Batch(
			m.app.Client().Init(),
			m.app.Dialog().Init(),
			m.titlebar.Init(),
			m.statusbar.Init(),
		),
		types.FocusChange(types.FocusPage),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Debug("tui update", "message", fmt.Sprintf("%#v", msg), "type", fmt.Sprintf("%T", msg))

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case types.AppQuitMsg:
		return m, tea.Sequence(
			tea.SetWindowTitle(""), // TODO: https://github.com/charmbracelet/bubbletea/issues/1474
			tea.Quit,
		)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		return m, tea.Batch(
			m.app.Page().Update(tea.WindowSizeMsg{
				Height: msg.Height - 2, // -1=titlebar, -1=statusbar.
				Width:  msg.Width,
			}),
			m.app.Dialog().Update(msg),
			m.titlebar.Update(tea.WindowSizeMsg{
				Height: 1,
				Width:  msg.Width,
			}),
			m.statusbar.Update(tea.WindowSizeMsg{
				Height: 1,
				Width:  msg.Width,
			}),
		)
	case tea.BackgroundColorMsg, tea.ColorProfileMsg:
		return m, styles.Theme.Update(msg)
	case tea.KeyMsg:
		// TODO: temporary bindings, will eventually be handled by the config package.
		switch {
		case msg.String() == "[":
			return m, styles.Theme.PreviousTint()
		case msg.String() == "]":
			return m, styles.Theme.NextTint()
		}

		switch m.focused {
		case types.FocusDialog:
			// If active dialog isn't the help dialog, and the help key is pressed,
			// open the help dialog.
			if v := m.app.Dialog().Get(false); v != nil && !v.HasInputFocus() {
				if _, ok := v.(*help.Model); !ok && key.Matches(msg, types.KeyHelp) {
					return m, types.OpenDialog(help.New(m.app))
				}
			}
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
	case debouncer.InvokeMsg, debouncer.DebounceMsg:
		return m, m.debouncer.Update(msg)
	}

	return m, tea.Batch(append(
		cmds,
		m.app.Page().Update(msg),
		m.app.Dialog().Update(msg),
		m.app.Client().Update(msg),
		m.titlebar.Update(msg),
		m.statusbar.Update(msg),
	)...)
}

func (m *Model) appTitle() string {
	switch {
	case m.focused == types.FocusDialog && m.app.Dialog().Len(false) > 0:
		d := m.app.Dialog().Get(false)
		if d != nil {
			return config.AppTitle(d.GetTitle())
		}
	case m.focused == types.FocusPage:
		return config.AppTitle(m.app.Page().Get().GetTitle())
	}
	return ""
}

func (m *Model) View() (view tea.View) {
	view.BackgroundColor = styles.Theme.AppBg()
	view.ForegroundColor = styles.Theme.AppFg()
	view.WindowTitle = m.appTitle()

	if m.width < MinWinWidth || m.height < MinWinHeight {
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center).
				Height(m.height).
				Width(m.width).
				Render(styles.IconCaution() + " window too small, resize")),
		)
		return view
	}

	view.Layer = lipgloss.NewCanvas(m.app.Dialog().SetLayers(
		lipgloss.NewLayer(
			lipgloss.NewStyle().
				Width(m.width).
				Height(m.height).
				Render(
					lipgloss.JoinVertical(
						lipgloss.Top,
						m.titlebar.View(),
						m.app.Page().View(),
						m.statusbar.View(),
					),
				),
		).ID("main").Z(0).X(0).Y(0),
	))

	return view
}
