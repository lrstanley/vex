// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package dialogselector

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/datatable"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type Listable interface {
	Suggestions() []string
	Len() int
	GetData() ([]string, [][]string)
}

type Config struct {
	List              Listable
	FilterPlaceholder string
	SelectFunc        func(row []string) tea.Cmd
}

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	*types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	config        Config
	previousInput string

	// Styles.
	BaseStyle  lipgloss.Style
	InputStyle lipgloss.Style

	// Child components.
	input textinput.Model
	table *datatable.Model[[]string]
}

func New(app types.AppState, config Config) *Model {
	m := &Model{
		ComponentModel: &types.ComponentModel{},
		app:            app,
		config:         config,
		input:          textinput.New(),
	}

	if m.config.FilterPlaceholder != "" {
		m.input.Placeholder = m.config.FilterPlaceholder
	} else {
		m.input.Placeholder = "type to filter"
	}
	m.input.VirtualCursor = true
	m.input.ShowSuggestions = true
	m.input.SetSuggestions(m.config.List.Suggestions())

	m.table = datatable.New(app, datatable.Config[[]string]{
		SelectFn: func(row []string) tea.Cmd {
			return config.SelectFunc(row)
		},
		RowFn: func(row []string) []string { return row },
	})

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.BaseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg())

	m.InputStyle = lipgloss.NewStyle().
		Padding(0, 1, 1, 1)

	m.input.Styles.Focused.Placeholder = m.input.Styles.Focused.Placeholder.
		Foreground(styles.Theme.Fg()).Faint(true)

	m.input.Styles.Focused.Suggestion = m.input.Styles.Focused.Suggestion.
		Foreground(styles.Theme.Fg()).Faint(true)

	m.input.Styles.Focused.Text = m.input.Styles.Focused.Text.
		Foreground(styles.Theme.Fg())

	m.input.Styles.Focused.Prompt = m.input.Styles.Focused.Prompt.
		Foreground(styles.Theme.Fg())

	m.input.Styles.Blurred.Prompt = m.input.Styles.Blurred.Prompt.
		Foreground(styles.Theme.InfoFg())

	m.input.Styles.Cursor.Color = styles.Theme.Fg()
	// TODO: bug with bubbles v2, returns cursor.BlinkMsg, then returns cursor.blinkCanceled,
	// which we can't handle because its private. Can technically use %T and strings.Contains,
	// but even that, the cursor stops blinking after the first blink, and disappears.
	m.input.Styles.Cursor.Blink = false
}

func (m *Model) updateDimensions() {
	m.InputStyle = m.InputStyle.Height(1)

	// TODO: https://github.com/charmbracelet/bubbles/issues/812
	m.input.SetWidth(m.Width - m.InputStyle.GetHorizontalFrameSize() - 5)
	m.InputStyle = m.InputStyle.Width(m.Width - m.InputStyle.GetHorizontalFrameSize())

	// Re-calculate the height so the dialog is only as big as we need, up to the max
	// of the default of [DialogModel.Size].
	m.Height = min(m.Height, m.InputStyle.GetVerticalFrameSize()+m.config.List.Len()+1) // +1=table header.

	m.table.Height = m.Height - m.InputStyle.GetVerticalFrameSize()
	m.table.Width = m.Width
}

func (m *Model) Value() string {
	return m.input.Value()
}

func (m *Model) Init() tea.Cmd {
	return m.input.Focus()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width
		m.updateDimensions()
		m.updateTable()
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		m.updateDimensions()
	case tea.KeyMsg:
		switch {
		case msg.String() == "up":
			if !m.table.Focused() {
				m.table.Focus()
				m.input.Blur()
			}
		case msg.String() == "down":
			if !m.table.Focused() {
				m.table.Focus()
				m.input.Blur()
			}
		case key.Matches(msg, types.KeySelectItem):
			selected, ok := m.table.GetSelectedData()
			if !ok {
				return nil
			}
			return m.config.SelectFunc(selected)
		case key.Matches(msg, types.KeyCancel):
			if m.input.Value() != "" {
				m.input.Reset()
				return nil
			}
			return nil
		default:
			if !m.input.Focused() {
				m.table.Blur()
				cmds = append(cmds, m.input.Focus())
			} else {
				if m.input.Focused() {
					m.input, cmd = m.input.Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		if m.input.Focused() {
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	var tableUpdated bool
	if m.input.Value() != m.previousInput {
		m.previousInput = m.input.Value()
		m.table.SetFilter(m.input.Value())
		tableUpdated = true
	}

	if m.table.Focused() || tableUpdated {
		cmds = append(cmds, m.table.Update(msg))
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateTable() {
	cols, rows := m.config.List.GetData()
	m.table.SetData(cols, rows)
}

func (m *Model) View() string {
	var out []string

	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	out = append(out, m.InputStyle.Render(m.input.View()))

	if m.table.FilteredDataLen() == 0 {
		out = append(out, lipgloss.NewStyle().Width(m.Width).
			Align(lipgloss.Center).
			Foreground(styles.Theme.ErrorFg()).
			Background(styles.Theme.ErrorBg()).
			Padding(0, 1).
			Render("no results found"),
		)
	} else {
		out = append(out, lipgloss.NewStyle().MaxHeight(m.table.FilteredDataLen()+1).Render(m.table.View()))
	}

	return m.BaseStyle.Render(lipgloss.JoinVertical(lipgloss.Top, out...))
}
