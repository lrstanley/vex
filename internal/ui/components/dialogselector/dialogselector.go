// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package dialogselector

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/table"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
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
	table table.Model
}

func New(app types.AppState, config Config) *Model {
	m := &Model{
		ComponentModel: &types.ComponentModel{},
		app:            app,
		config:         config,
		input:          textinput.New(),
		table:          table.New(),
	}

	if m.config.FilterPlaceholder != "" {
		m.input.Placeholder = m.config.FilterPlaceholder
	} else {
		m.input.Placeholder = "type to filter"
	}
	m.input.VirtualCursor = true
	m.input.ShowSuggestions = true
	m.input.SetSuggestions(m.config.List.Suggestions())

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.BaseStyle = lipgloss.NewStyle().
		Background(styles.Theme.Bg()).
		Foreground(styles.Theme.DialogFg())

	m.InputStyle = lipgloss.NewStyle().
		Padding(0, 1, 1, 1)

	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	m.table.SetStyles(s)
}

func (m *Model) updateDimensions() {
	m.InputStyle = m.InputStyle.Height(1)

	// TODO: https://github.com/charmbracelet/bubbles/issues/812
	m.input.SetWidth(m.Width - m.InputStyle.GetHorizontalFrameSize() - 5)
	m.InputStyle = m.InputStyle.Width(m.Width - m.InputStyle.GetHorizontalFrameSize())

	// Re-calculate the height so the dialog is only as big as we need, up to the max
	// of the default of [DialogModel.Size].
	m.Height = min(m.Height, m.InputStyle.GetVerticalFrameSize()+m.config.List.Len()+1) // +1=table header.

	m.table.SetWidth(m.Width)
	m.table.SetHeight(m.Height - m.InputStyle.GetVerticalFrameSize())
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
			row := m.table.SelectedRow()
			if row == nil {
				return nil
			}

			return m.config.SelectFunc(row)
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
		m.updateTable()
		tableUpdated = true
	}

	if m.table.Focused() || tableUpdated {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateTable() {
	cols, rows := m.config.List.GetData()

	// Filter first.
	filter := m.input.Value()
	var trows []table.Row
	for i := range rows {
		if filter != "" {
			if !strings.Contains(strings.ToLower(strings.Join(rows[i], ":")), strings.ToLower(filter)) {
				continue
			}
		}
		trows = append(trows, rows[i])
	}

	// Since the table component doesn't really give us an easy way to know how much
	// padding it will have, but we want to have the last column take up the remaining
	// space, we need to calculate the width of each column, and then set the last column
	// to the remaining width.
	colWidths := make([]int, len(cols))
	for i := range cols {
		colWidths[i] = lipgloss.Width(cols[i])
		for _, row := range trows {
			colWidths[i] = max(colWidths[i], lipgloss.Width(row[i]))
		}
	}
	colWidths[len(colWidths)-1] = m.Width - 2 // -2=table padding.
	if len(cols) > 1 {
		for i := range len(colWidths) - 1 {
			colWidths[len(colWidths)-1] -= colWidths[i] + 2 // +2=column padding.
		}
	}

	tcols := make([]table.Column, len(cols))
	for i := range cols {
		tcols[i].Title = cols[i]
		tcols[i].Width = colWidths[i]
	}

	m.table.SetColumns(tcols)
	m.table.SetRows(trows)
}

func (m *Model) View() string {
	var out []string

	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	out = append(out, m.InputStyle.Render(m.input.View()))

	if len(m.table.Rows()) == 0 {
		out = append(out, lipgloss.NewStyle().Width(m.Width).
			Align(lipgloss.Center).
			Foreground(styles.Theme.ErrorFg()).
			Background(styles.Theme.ErrorBg()).
			Padding(0, 1).
			Render("no results found"),
		)
	} else {
		out = append(out, lipgloss.NewStyle().MaxHeight(len(m.table.Rows())+1).Render(m.table.View()))
	}

	return lipgloss.JoinVertical(lipgloss.Top, out...)
}
