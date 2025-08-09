// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package datatable

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/bubbles/v2/table"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/fuzzy"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/loader"
	"github.com/lrstanley/vex/internal/ui/styles"
)

// TableStyles contains all the customizable styles for the table component.
type TableStyles struct {
	Base      lipgloss.Style
	NoResults lipgloss.Style
	Table     table.Styles
}

// Config contains the configuration for the table component.
type Config[T any] struct {
	NoResultsMsg       string
	NoResultsFilterMsg string
	FilterFunc         func(filter string, values []T) []T
	SelectFn           func(value T) tea.Cmd
	RowFn              func(value T) []string
	FetchFn            func() tea.Cmd
}

var _ types.Component = (*Model[any])(nil) // Ensure we implement the component interface.

type Model[T any] struct {
	*types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	config   Config[T]
	filter   string
	columns  []string
	data     []T
	filtered []T
	selected *T
	loading  bool

	// Styles.
	styles TableStyles

	// Child components.
	table  table.Model
	loader *loader.Model
}

// New returns a new table component. it will by default be in a loading state.
func New[T any](app types.AppState, config Config[T]) *Model[T] {
	m := &Model[T]{
		ComponentModel: &types.ComponentModel{},
		app:            app,
		config:         config,
		loading:        true,
		table:          table.New(table.WithFocused(true)),
		loader:         loader.New(),
	}

	if m.config.NoResultsMsg == "" {
		m.config.NoResultsMsg = "no results found"
	}
	if m.config.NoResultsFilterMsg == "" {
		m.config.NoResultsFilterMsg = "no results found for %q"
	}

	if m.config.FilterFunc == nil {
		m.config.FilterFunc = func(filter string, values []T) []T {
			return fuzzy.FindRankedRow(filter, values, m.config.RowFn)
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
	return m.SetLoading()
}

// Fetch fetches the data from the config.FetchFn.
func (m *Model[T]) Fetch(setLoading bool) tea.Cmd {
	if m.config.FetchFn == nil {
		return nil
	}
	cmds := []tea.Cmd{m.config.FetchFn()}
	if setLoading {
		cmds = append(cmds, m.SetLoading())
	}
	return tea.Batch(cmds...)
}

// GetData returns the data.
func (m *Model[T]) GetData() []T {
	return m.data
}

// GetFilteredData returns the filtered data.
func (m *Model[T]) GetFilteredData() []T {
	return m.filtered
}

// DataLen returns the total number of data entries.
func (m *Model[T]) DataLen() int {
	return len(m.data)
}

// FilteredDataLen returns the number of filtered rows.
func (m *Model[T]) FilteredDataLen() int {
	return len(m.filtered)
}

func (m *Model[T]) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetDimensions(msg.Width, msg.Height)
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		m.updateTable()
	case tea.KeyMsg:
		switch {
		case (key.Matches(msg, types.KeySelectItem) || key.Matches(msg, types.KeySelectItemAlt)) && m.table.Focused():
			if m.config.SelectFn != nil {
				selected, ok := m.GetSelectedData()
				if ok {
					cmds = append(cmds, m.config.SelectFn(selected))
				}
			}
		}
	case spinner.TickMsg:
		if m.loader.SpinnerID() != msg.ID || !m.loading {
			return nil
		}
		return m.loader.Update(msg)
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(append(
		cmds,
		m.loader.Update(msg),
	)...)
}

func (m *Model[T]) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
	m.loader.SetHeight(height)
	m.loader.SetWidth(width)
	m.updateDimensions()
	m.updateTable()
	m.table.GotoTop()
}

func (m *Model[T]) SetWidth(width int) {
	m.SetDimensions(width, m.Height)
}

func (m *Model[T]) SetHeight(height int) {
	m.SetDimensions(m.Width, height)
}

func (m *Model[T]) updateTable() {
	m.filtered = m.config.FilterFunc(m.filter, m.data)

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
		trows[i] = m.config.RowFn(row)
	}
	m.table.SetColumns(tcols)
	m.table.SetRows(trows)
}

func (m *Model[T]) calculateColumnWidths() []int {
	if len(m.columns) == 0 {
		return []int{}
	}

	colWidths := make([]int, len(m.columns))

	// Calculate minimum width needed for each column.
	for i := range m.columns {
		colWidths[i] = ansi.StringWidth(m.columns[i])
		for _, data := range m.filtered {
			row := m.config.RowFn(data)
			if i < len(row) {
				colWidths[i] = max(colWidths[i], ansi.StringWidth(row[i]))
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

// Focus focuses the table.
func (m *Model[T]) Focus() {
	m.table.Focus()
}

// Blur unfocuses the table.
func (m *Model[T]) Blur() {
	m.table.Blur()
}

// Focused returns true if the table is focused.
func (m *Model[T]) Focused() bool {
	return m.table.Focused()
}

// GoToTop scrolls the table to the top.
func (m *Model[T]) GoToTop() {
	m.table.GotoTop()
}

// GotoBottom scrolls the table to the bottom.
func (m *Model[T]) GotoBottom() {
	m.table.GotoBottom()
}

// SetIndex sets the cursor (selected row) to the given index.
func (m *Model[T]) SetIndex(i int) {
	m.table.SetCursor(max(0, min(i, len(m.filtered)-1))) // Clamped to 0-len(m.filtered)-1.
}

// SetFilter sets the filter string and updates the table. Setting to an empty
// string will render all rows.
func (m *Model[T]) SetFilter(filter string) {
	if m.filter != filter {
		m.filter = filter
		m.table.GotoTop()
	} else {
		m.filter = filter
	}
	m.updateTable()
}

// SetData sets the data for the table.
func (m *Model[T]) SetData(columns []string, values []T) {
	oldLen := len(m.data)
	wasNil := m.data == nil

	m.columns = columns
	m.data = make([]T, 0, len(values))
	m.data = append(m.data, values...)
	m.updateTable()
	m.loading = false

	// If we might be out of the view, or the data was originally nil (initial load),
	// then we need to go to the top.
	if oldLen > len(m.data) || wasNil {
		m.table.GotoTop()
	}
}

// GetSelectedData returns the selected data and a boolean indicating if it was
// found and valid.
func (m *Model[T]) GetSelectedData() (T, bool) {
	var v T
	row := m.table.SelectedRow()
	if row == nil {
		return v, false
	}

	selectedIndex := m.table.Cursor()
	if selectedIndex >= 0 && selectedIndex < len(m.filtered) {
		return m.filtered[selectedIndex], true
	}

	return v, false
}

// SetLoading sets the loading state to true. Will automatically be set back to
// false once data has been updated/fetched/etc.
func (m *Model[T]) SetLoading() tea.Cmd {
	m.loading = true
	return m.loader.Active()
}

func (m *Model[T]) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	var out []string

	switch {
	case m.loading:
		out = append(out, m.loader.View())
	case len(m.data) == 0:
		out = append(out, m.styles.NoResults.
			Width(m.Width).
			Render(m.config.NoResultsMsg))
	case len(m.filtered) == 0 && m.filter != "":
		out = append(out, m.styles.NoResults.
			Width(m.Width).
			Render(fmt.Sprintf(m.config.NoResultsFilterMsg, m.filter)))
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
