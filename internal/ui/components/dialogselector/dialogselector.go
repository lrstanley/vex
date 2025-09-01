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
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type Styles struct {
	Base      lipgloss.Style
	InputBase lipgloss.Style
	Input     textinput.Styles
}

type Config struct {
	Columns           []*table.Column
	FilterPlaceholder string
	SelectFunc        func(id string) tea.Cmd
}

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	*types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	config        Config
	previousInput string
	suggestions   []string
	items         [][]string

	// Styles.
	styles         Styles
	providedStyles Styles

	// Child components.
	input textinput.Model
	table *table.Model[*table.StaticRow[[]string]]
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
	m.input.SetVirtualCursor(true)
	m.input.ShowSuggestions = true

	m.table = table.New(app, config.Columns, table.Config[*table.StaticRow[[]string]]{
		SelectFn: func(row *table.StaticRow[[]string]) tea.Cmd {
			return config.SelectFunc(string(row.ID()))
		},
		RowFn: func(row *table.StaticRow[[]string]) []string { return row.Value },
	})

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.styles.Base = m.providedStyles.Base.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()),
		)

	m.styles.InputBase = m.providedStyles.InputBase.
		Padding(0, 1, 1, 1)

	inputStyles := m.providedStyles.Input

	inputStyles.Focused.Placeholder = inputStyles.Focused.Placeholder.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()).
				Faint(true),
		)

	inputStyles.Focused.Suggestion = inputStyles.Focused.Suggestion.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()).
				Faint(true),
		)

	inputStyles.Focused.Text = inputStyles.Focused.Text.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()),
		)

	inputStyles.Focused.Prompt = inputStyles.Focused.Prompt.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()),
		)

	inputStyles.Blurred.Prompt = inputStyles.Blurred.Prompt.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.InfoFg()),
		)

	inputStyles.Cursor.Color = styles.Theme.AppFg()
	// TODO: bug with bubbles v2, returns cursor.BlinkMsg, then returns cursor.blinkCanceled,
	// which we can't handle because its private. Can technically use %T and strings.Contains,
	// but even that, the cursor stops blinking after the first blink, and disappears.
	inputStyles.Cursor.Blink = false

	m.input.SetStyles(inputStyles)
}

func (m *Model) SetStyles(s Styles) {
	m.providedStyles = s
	m.initStyles()
}

func (m *Model) SetSuggestions(suggestions []string) {
	m.suggestions = suggestions
	m.input.SetSuggestions(suggestions)
}

func (m *Model) SetItems(items [][]string) {
	m.items = items
	m.updateTable()
}

func (m *Model) updateTable() {
	var out [][]string
	var ids []table.ID
	for _, row := range m.items {
		out = append(out, row[1:])
		ids = append(ids, table.ID(row[0]))
	}

	var i int
	m.table.SetRows(table.RowsFrom(out, func(row []string) table.ID {
		id := ids[i]
		i++
		return id
	}))
}

func (m *Model) updateDimensions() {
	m.styles.InputBase = m.styles.InputBase.Height(1)

	// TODO: https://github.com/charmbracelet/bubbles/issues/812
	m.input.SetWidth(m.Width - m.styles.InputBase.GetHorizontalFrameSize() - 5)
	m.styles.InputBase = m.styles.InputBase.Width(m.Width - m.styles.InputBase.GetHorizontalFrameSize())

	// Re-calculate the height so the dialog is only as big as we need, up to the max
	// of the default of [DialogModel.Size].
	m.Height = min(m.Height, m.styles.InputBase.GetVerticalFrameSize()+len(m.items)+1) // +1=table header.

	m.table.Height = m.Height - m.styles.InputBase.GetVerticalFrameSize()
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
		return m.table.Update(msg)
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		m.updateDimensions()
	case tea.KeyMsg:
		switch {
		case msg.String() == "up" || msg.String() == "down":
			// Move down in table - the table handles this internally.
			return m.table.Update(msg)
		case msg.String() == "left" || msg.String() == "right":
			// Move left/right in input - the input handles this internally.
			m.input, cmd = m.input.Update(msg)
			return cmd
		case key.Matches(msg, types.KeySelectItem):
			selected, ok := m.table.GetSelectedRow()
			if !ok {
				return nil
			}
			return m.config.SelectFunc(string(selected.ID()))
		case key.Matches(msg, types.KeyCancel):
			if m.input.Value() != "" {
				m.input.Reset()
				return nil
			}
			return nil
		default:
			// Handle input focus
			if !m.input.Focused() {
				cmds = append(cmds, m.input.Focus())
			} else if m.input.Focused() {
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		if m.input.Focused() {
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update filter if input changed
	if m.input.Value() != m.previousInput {
		m.previousInput = m.input.Value()
		m.table.SetFilter(m.input.Value())
	}

	// Always update the table to handle navigation
	cmds = append(cmds, m.table.Update(msg))

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	var out []string

	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	out = append(out, m.styles.InputBase.Render(m.input.View()))

	if m.table.TotalFilteredRows() == 0 {
		out = append(out, lipgloss.NewStyle().Width(m.Width).
			Align(lipgloss.Center).
			Foreground(styles.Theme.ErrorFg()).
			Background(styles.Theme.ErrorBg()).
			Padding(0, 1).
			Render("no results found"),
		)
	} else {
		out = append(out, lipgloss.NewStyle().MaxHeight(m.table.TotalFilteredRows()+1).Render(m.table.View()))
	}

	return m.styles.Base.Render(lipgloss.JoinVertical(lipgloss.Top, out...))
}
