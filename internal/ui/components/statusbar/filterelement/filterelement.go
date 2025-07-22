// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package filterelement

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/colors"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const (
	filterMaxWidth = 35
	filterPrefix   = "filter: "
)

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	debounce      types.Debouncer
	inputWidth    int
	previousValue string
	previousUUID  string

	// Styles.
	baseStyle   lipgloss.Style
	prefixStyle lipgloss.Style

	// Child components.
	filter textinput.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		filter:         textinput.New(),
	}

	m.filter.Placeholder = "type to filter..."
	m.filter.VirtualCursor = true
	m.filter.Prompt = ""
	m.setStyles()
	m.filter.KeyMap.Paste = key.NewBinding(key.WithKeys("ctrl+v", "ctrl+shift+v"))
	m.inputWidth = filterMaxWidth - lipgloss.Width(filterPrefix) - m.prefixStyle.GetHorizontalFrameSize()
	m.filter.SetWidth(m.inputWidth)
	return m
}

func (m *Model) setStyles() {
	m.baseStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		Background(styles.Theme.StatusBarFilterBg())

	m.prefixStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.StatusBarFilterFg()).
		Background(styles.Theme.StatusBarFilterBg())

	m.filter.Styles.Focused.Placeholder = m.filter.Styles.Focused.Placeholder.
		Foreground(colors.Darken(styles.Theme.StatusBarFilterFg(), 20)).
		Background(colors.Darken(styles.Theme.StatusBarFilterBg(), 10))

	m.filter.Styles.Focused.Text = m.filter.Styles.Focused.Text.
		Foreground(styles.Theme.Tint().White).
		Background(colors.Darken(styles.Theme.StatusBarFilterBg(), 10))
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	case types.AppFocusChangedMsg:
		if msg.ID != types.FocusStatusBar {
			m.filter.Blur()
		} else {
			cmds = append(cmds, m.filter.Focus())
		}

		if !m.app.Page().Get().GetSupportFiltering() || m.previousUUID != m.app.Page().Get().UUID() {
			m.filter.Reset()
		}

		cmds = append(cmds, m.sendFilter())

		return tea.Batch(cmds...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel) && m.filter.Value() != "":
			m.filter.Reset()
			return m.sendFilter()
		case key.Matches(msg, types.KeyCancel) && m.filter.Value() == "":
			return types.RequestPreviousFocus()
		case key.Matches(msg, types.KeyFilter) && m.filter.Value() == "":
			return nil
		case key.Matches(msg, types.KeySelectItem):
			return tea.Batch(
				types.RequestPreviousFocus(),
				m.sendFilter(),
			)
		default:
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			cmds = append(cmds, cmd, m.debounce.Send())
		}
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		cmds = append(cmds, cmd, m.debounce.Send())
	case types.DebounceMsg:
		if m.debounce.Is(msg) && m.filter.Value() != m.previousValue {
			m.previousValue = m.filter.Value()
			cmds = append(cmds, m.sendFilter())
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) sendFilter() tea.Cmd {
	m.previousUUID = m.app.Page().Get().UUID()
	return types.AppFilter(m.previousUUID, m.filter.Value())
}

func (m *Model) View() string {
	return m.baseStyle.Render(m.prefixStyle.Render(filterPrefix) + m.filter.View())
}
