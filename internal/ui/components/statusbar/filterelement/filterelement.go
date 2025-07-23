// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package filterelement

import (
	"slices"

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
	debounce     types.Debouncer
	inputWidth   int
	previousUUID string
	filterState  map[string]string // page uuid -> filter value.

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
		filterState:    make(map[string]string),
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
		Foreground(styles.Theme.StatusBarFilterTextFg()).
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

		currentUUID := m.app.Page().Get().UUID()

		// Save current filter state before switching.
		if m.previousUUID != "" && m.previousUUID != currentUUID {
			m.filterState[m.previousUUID] = m.filter.Value()
		}

		// Remove filter state for pages that are no longer active.
		activeUUIDs := m.app.Page().UUIDs()
		for uuid := range m.filterState {
			if !slices.Contains(activeUUIDs, uuid) {
				delete(m.filterState, uuid)
			}
		}

		// Update to new page.
		if m.previousUUID != currentUUID {
			m.previousUUID = currentUUID

			// Restore filter state for the new page.
			if savedFilter, exists := m.filterState[currentUUID]; exists {
				m.filter.SetValue(savedFilter)
			} else {
				m.filter.Reset()
			}
		}

		// Always re-send the current filter state for the active page.
		cmds = append(cmds, m.sendFilter())

		return tea.Batch(cmds...)
	case types.AppFilterClearedMsg:
		if _, ok := m.filterState[m.app.Page().Get().UUID()]; ok {
			m.filter.Reset()
			delete(m.filterState, m.app.Page().Get().UUID())
			return m.sendFilter()
		}
		return nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel) && m.filter.Value() != "":
			m.filter.Reset()
			delete(m.filterState, m.app.Page().Get().UUID())
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
		if m.debounce.Is(msg) {
			// Update the filter state for the current page.
			currentUUID := m.app.Page().Get().UUID()
			m.filterState[currentUUID] = m.filter.Value()
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
