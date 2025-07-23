// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/table"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type mockRow struct{}

func (r mockRow) Get() mockRow {
	return r
}

func (r mockRow) Row() []string {
	return []string{"mock"}
}

type Row[T any] interface {
	Get() T
	Row() []string
}

// TableStyles contains all the customizable styles for the table component.
type TableStyles struct {
	Base      lipgloss.Style
	NoResults lipgloss.Style
	Loading   lipgloss.Style
	Table     table.Styles
}

// Config contains the configuration for the table component.
type Config[T any] struct {
	NoResultsMsg string
	FilterFunc   func(item T, filter string) bool
	OnSelect     func(item T)
}

var _ types.Component = (*Model[mockRow])(nil) // Ensure we implement the component interface.

type Model[T Row[T]] struct {
	*types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	config   Config[T]
	filter   string
	columns  []string
	allData  []T
	filtered []T
	loading  bool

	// Styles.
	styles TableStyles

	// Child components.
	table   table.Model
	spinner spinner.Model
}

// New returns a new table component. it will by default be in a loading state.
func New[T Row[T]](app types.AppState, config Config[T]) *Model[T] {
	m := &Model[T]{
		ComponentModel: &types.ComponentModel{},
		app:            app,
		config:         config,
		loading:        true,
		table:          table.New(table.WithFocused(true)),
		spinner:        spinner.New(),
	}

	m.spinner.Spinner = spinner.MiniDot

	if m.config.NoResultsMsg == "" {
		m.config.NoResultsMsg = "no results found"
	}

	if m.config.FilterFunc == nil {
		m.config.FilterFunc = func(item T, filter string) bool {
			if filter == "" {
				return true
			}
			return strings.Contains(strings.ToLower(strings.Join(item.Row(), " ")), strings.ToLower(filter))
		}
	}

	m.styles.Table = table.DefaultStyles()
	m.initStyles()
	m.updateTable()
	return m
}

func (m *Model[T]) initStyles() {
	m.styles.Base = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg())

	m.styles.NoResults = lipgloss.NewStyle().
		Foreground(styles.Theme.ErrorFg()).
		Background(styles.Theme.ErrorBg()).
		Padding(0, 1).
		Align(lipgloss.Center)

	m.styles.Loading = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg()).
		Padding(0, 1).
		Align(lipgloss.Center)

	m.spinner.Style = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg())

	m.styles.Table.Header = m.styles.Table.Header.Bold(true).
		Foreground(styles.Theme.Fg())

	m.styles.Table.Selected = m.styles.Table.Selected.
		Foreground(styles.Theme.InfoFg()).
		Background(styles.Theme.InfoBg()).
		Bold(true)

	m.table.SetStyles(m.styles.Table)
}

func (m *Model[T]) SetStyles(styles TableStyles) {
	m.styles = styles
	m.table.SetStyles(m.styles.Table)
}

func (m *Model[T]) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model[T]) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width
		m.updateDimensions()
		m.updateTable()
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		m.updateTable()
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		switch {
		case (key.Matches(msg, types.KeySelectItem) || key.Matches(msg, types.KeySelectItemAlt)) && m.table.Focused():
			selected := m.GetSelectedData()
			if selected != nil && m.config.OnSelect != nil {
				m.config.OnSelect(*selected)
			}
		case key.Matches(msg, types.KeyCancel):
			return nil
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model[T]) updateTable() {
	m.filtered = make([]T, 0, len(m.allData))

	if m.filter != "" && m.config.FilterFunc != nil {
		for _, item := range m.allData {
			if m.config.FilterFunc(item, m.filter) {
				m.filtered = append(m.filtered, item)
			}
		}
	} else {
		m.filtered = m.allData
	}

	// Calculate column widths.
	colWidths := m.calculateColumnWidths()

	// Create table columns.
	tcols := make([]table.Column, len(m.columns))
	for i := range m.columns {
		tcols[i].Title = m.columns[i]
		tcols[i].Width = colWidths[i]
	}

	trows := make([]table.Row, len(m.filtered))
	for i, row := range m.filtered {
		trows[i] = row.Row()
	}
	m.table.SetColumns(tcols)
	m.table.SetRows(trows)
	m.table.GotoTop()
}

func (m *Model[T]) calculateColumnWidths() []int {
	if len(m.columns) == 0 {
		return []int{}
	}

	colWidths := make([]int, len(m.columns))

	// Calculate minimum width needed for each column.
	for i := range m.columns {
		colWidths[i] = lipgloss.Width(m.columns[i])
		for _, data := range m.filtered {
			row := data.Row()
			if i < len(row) {
				colWidths[i] = max(colWidths[i], lipgloss.Width(row[i]))
			}
		}
	}

	// Check if total width exceeds available width.
	totalWidth := 0
	for _, width := range colWidths {
		totalWidth += width + 2 // +2 for column padding
	}
	totalWidth -= 2 // Remove padding from last column

	if totalWidth > m.Width-2 { // -2 for table padding
		// Set each column to its maximum potential size.
		return colWidths
	}

	// Auto-calculate: set last column to remaining width.
	colWidths[len(colWidths)-1] = m.Width - 2 // -2=table padding
	if len(m.columns) > 1 {
		for i := range len(colWidths) - 1 {
			colWidths[len(colWidths)-1] -= colWidths[i] + 2 // +2=column padding
		}
	}

	return colWidths
}

func (m *Model[T]) updateDimensions() {
	m.table.SetWidth(m.Width)
	m.table.SetHeight(m.Height)
}

func (m *Model[T]) Focus() {
	m.table.Focus()
}

func (m *Model[T]) Blur() {
	m.table.Blur()
}

func (m *Model[T]) SetFilter(filter string) {
	m.filter = filter
	m.updateTable()
}

func (m *Model[T]) SetData(columns []string, data []T) {
	m.columns = columns
	m.allData = data
	m.updateTable()
	m.loading = false
}

func (m *Model[T]) GetSelectedData() *T {
	row := m.table.SelectedRow()
	if row == nil {
		return nil
	}

	selectedIndex := m.table.Cursor()
	if selectedIndex >= 0 && selectedIndex < len(m.filtered) {
		return &m.filtered[selectedIndex]
	}

	return nil
}

func (m *Model[T]) SetLoading() tea.Cmd {
	m.loading = true
	return m.spinner.Tick
}

func (m *Model[T]) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	var out []string

	switch {
	case m.loading:
		centeredLoading := m.styles.Loading.
			Width(m.Width).
			Height(m.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.spinner.View() + m.styles.Loading.Render("loading..."))

		out = append(out, centeredLoading)
	case len(m.filtered) == 0:
		out = append(out, m.styles.NoResults.
			Width(m.Width).
			Render(m.config.NoResultsMsg))
	default:
		out = append(out, m.styles.Base.
			MaxHeight(m.Height).
			MaxWidth(m.Width).
			Render(m.table.View()))
	}

	return m.styles.Base.
		Height(m.Height).
		Width(m.Width).
		Render(lipgloss.JoinVertical(lipgloss.Top, out...))
}
