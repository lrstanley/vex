// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/loader"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type Styles struct {
	Base                      lipgloss.Style
	NoResults                 lipgloss.Style
	Header                    lipgloss.Style
	Cell                      lipgloss.Style
	SelectedRow               lipgloss.Style
	HighlightedRow            lipgloss.Style
	HighlightedAndSelectedRow lipgloss.Style
}

type ID string

type Config[T Row] struct {
	SelectFn           func(value T) tea.Cmd
	RowFn              func(value T) []string
	FetchFn            func() tea.Cmd
	NoResultsMsg       string
	NoResultsFilterMsg string
	AllowHighlighting  bool
}

type exampleRow struct{}

func (e exampleRow) ID() ID {
	return "example"
}

var _ types.Component = (*Model[exampleRow])(nil) // Ensure we implement the component interface.

type Model[T Row] struct {
	*types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	config Config[T]

	// filter is the current filter string.
	filter string

	// columns contains all of the columns (enabled or disabled).
	columns []*Column
	// columnIDMap is a map of column IDs to their index in the columns slice.
	columnIDMap map[ID]int

	// maxColumnWidths contains all of the enable columns max potential widths. It
	// does not include frame sizes for each cell.
	maxColumnWidths map[ID]int

	// data contains all of the rows.
	data []T
	// dataIDMap is a map of row IDs to their index in the data slice.
	dataIDMap map[ID]int

	// filtered contains all of the row IDs that match the current filter.
	filtered []ID

	// highlighted contains all of the row IDs that are highlighted.
	highlighted []ID

	// loading is true if the table is loading.
	loading bool

	// selectedIndex is the index of the currently selected row.
	selectedIndex int

	// yoffset is the offset of the view relative to the Y axis.
	yoffset int

	// xoffset is the offset of the view relative to the X axis.
	xoffset int

	// Styles.
	providedStyles Styles
	styles         Styles

	// Child components.
	loader *loader.Model
}

// New returns a new table component. it will by default be in a loading state.
func New[T Row](app types.AppState, columns []*Column, config Config[T]) *Model[T] {
	m := &Model[T]{
		ComponentModel:  &types.ComponentModel{},
		app:             app,
		config:          config,
		maxColumnWidths: make(map[ID]int),
		loading:         true,
		loader:          loader.New(),
	}

	if m.config.NoResultsMsg == "" {
		m.config.NoResultsMsg = "no results found"
	}
	if m.config.NoResultsFilterMsg == "" {
		m.config.NoResultsFilterMsg = "no results found for %q"
	}

	m.SetColumns(columns)
	m.initStyles()
	return m
}

func (m *Model[T]) initStyles() {
	m.styles.Base = m.providedStyles.Base.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()),
		)

	m.styles.NoResults = m.providedStyles.NoResults.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.ErrorFg()).
				Background(styles.Theme.ErrorBg()),
		).
		Padding(0, 1).
		Align(lipgloss.Center)

	m.styles.Cell = m.providedStyles.Cell.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()),
		).
		Padding(0, 1).
		UnsetMarginTop().
		UnsetMarginBottom().
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()

	m.styles.Header = m.providedStyles.Header.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()).
				Bold(true),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderLeft().
		UnsetBorderRight()

	m.styles.SelectedRow = m.providedStyles.SelectedRow.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()).
				Background(styles.Theme.InfoBg()).
				Bold(true),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()

	m.styles.HighlightedRow = m.providedStyles.HighlightedRow.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.SuccessFg()).
				Background(styles.Theme.SuccessBg()),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()

	m.styles.HighlightedAndSelectedRow = m.providedStyles.HighlightedAndSelectedRow.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.SuccessFg()).
				Background(styles.Theme.InfoBg()).
				Bold(true),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()
}

func (m *Model[T]) SetStyles(s Styles) {
	m.providedStyles = s
	m.initStyles()
}

func (m *Model[T]) Init() tea.Cmd {
	return m.SetLoading()
}

func (m *Model[T]) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width == 0 || msg.Height == 0 {
			return nil
		}
		m.SetDimensions(msg.Width, msg.Height)
		m.loader.SetDimensions(
			msg.Width-m.styles.Base.GetHorizontalFrameSize(),
			msg.Height-m.styles.Base.GetVerticalFrameSize(),
		)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		m.updateCalculations()

		cmds = append(cmds, m.loader.Update(msg))
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeySelectItem):
			if m.config.SelectFn != nil {
				selected, ok := m.GetSelectedRow()
				if ok {
					return m.config.SelectFn(selected)
				}
			}
			return nil
		case key.Matches(msg, types.KeySelectItemAlt):
			if !m.config.AllowHighlighting {
				return nil
			}

			selected, ok := m.GetSelectedRow()
			if !ok {
				return nil
			}
			if i := slices.Index(m.highlighted, selected.ID()); i != -1 {
				m.highlighted = slices.Delete(m.highlighted, i, i+1)
			} else {
				m.highlighted = append(m.highlighted, selected.ID())
			}
			return nil
		case key.Matches(msg, types.KeyUp):
			m.MoveUp(1)
		case key.Matches(msg, types.KeyDown):
			m.MoveDown(1)
		case key.Matches(msg, types.KeyLeft):
			m.MoveLeft(1)
		case key.Matches(msg, types.KeyRight):
			m.MoveRight(1)
		case key.Matches(msg, types.KeyPageUp):
			m.MoveUp(m.maxInnerTableHeight())
		case key.Matches(msg, types.KeyPageDown):
			m.MoveDown(m.maxInnerTableHeight())
		case key.Matches(msg, types.KeyGoToTop):
			m.GoToTop()
		case key.Matches(msg, types.KeyGoToBottom):
			m.GoToBottom()
		}
		return nil
	case spinner.TickMsg:
		if m.loader.SpinnerID() != msg.ID || !m.loading {
			return nil
		}
		return m.loader.Update(msg)
	}

	return tea.Batch(cmds...)
}

// SetFilter sets the filter string and updates the table. Setting to an empty
// string will clear all filtering.
func (m *Model[T]) SetFilter(filter string) {
	if m.filter == filter {
		return
	}

	selected, hasSelected := m.GetSelectedRow()

	m.filter = filter
	m.applyFiltering()

	// Try and re-select the same row as before filtering was applied.
	if hasSelected {
		m.SetSelected(selected.ID())
	}

	m.sanitizeHighlighted()
	m.updateCalculations()
}

// SetLoading returns a [tea.Cmd] that sets the loading state to true. Will
// automatically be set back to false once data has been updated/fetched/etc using
// [Model.SetRows].
func (m *Model[T]) SetLoading() tea.Cmd {
	m.loading = true
	return m.loader.Active()
}

// renderHeader renders the header of the table.
func (m *Model[T]) renderHeader(maxWidth int, hasScrollbar bool) string {
	out := make([]string, 0, len(m.columns))
	available := maxWidth
	xoffset := m.xoffset

	if hasScrollbar {
		available++
	}

	// If we have an offset, add an ellipsis to the left side, accounting for borders,
	// and other dynamic height.
	if m.xoffset > 0 {
		out = append(out,
			strings.Join(slices.Repeat(
				[]string{m.styles.Header.Inline(true).Render(formatter.TruncateEllipsis)},
				m.styles.Header.GetVerticalFrameSize()+1,
			), "\n"),
		)
		available--
	}

	hhfs := m.styles.Header.GetHorizontalFrameSize()
	hhvs := m.styles.Header.GetVerticalFrameSize()

	var s string
	var w, neww, cw int
	for i := range m.columns {
		if m.columns[i].Disabled {
			continue
		}

		cw = m.maxColumnWidths[m.columns[i].ID]

		// If the table has a scrollbar, and this is the last column, then we can actually
		// use the extra space above the scrollbar, specifically when it's a header.
		if hasScrollbar && i == len(m.columns)-1 {
			cw++
		}

		s = m.styles.Header.
			Height(1 + hhvs).
			Width(cw + hhfs).
			Align(m.columns[i].Align).
			Render(formatter.Trunc(m.columns[i].Title, cw))

		w = ansi.StringWidth(strings.Split(s, "\n")[0]) // Split to account for multiline headers.

		// Left side truncation.
		if xoffset > 0 {
			s = formatter.CutMultiline(s, xoffset+1, w)
			neww = ansi.StringWidth(strings.Split(s, "\n")[0])
			xoffset -= w - neww
			w = neww
		}

		// Right side truncation.
		if w > available {
			s = formatter.TruncMultiline(s, available)
		}
		available -= w

		out = append(out, s)

		if available <= 0 {
			break
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}

// renderBody renders the body of the table.
func (m *Model[T]) renderBody(maxWidth int) string {
	out := make([]string, 0, m.Height)
	selected, hasSelected := m.GetSelectedRow()

	var values, rowOut []string
	var s string
	var style lipgloss.Style
	var w, neww, cw, available, xoffset int
	for row := range m.GetVisibleRows() { // TODO: this needs to get ONLY active rows.
		rowOut = rowOut[:0]
		available = maxWidth
		xoffset = m.xoffset
		values = m.config.RowFn(row)

		if m.xoffset > 0 {
			rowOut = append(rowOut, formatter.TruncateEllipsis)
			available--
		}

		for i := range m.columns {
			if m.columns[i].Disabled {
				continue
			}

			cw = m.maxColumnWidths[m.columns[i].ID]

			switch {
			case slices.Contains(m.highlighted, row.ID()) && hasSelected && selected.ID() == row.ID():
				style = m.styles.HighlightedAndSelectedRow
			case slices.Contains(m.highlighted, row.ID()):
				style = m.styles.HighlightedRow
			case hasSelected && selected.ID() == row.ID(): // TODO: need something to cover both highlighted and selected.
				style = m.styles.SelectedRow
			default:
				style = m.styles.Cell
			}

			s = style.
				Height(1).
				Width(cw + m.styles.Cell.GetHorizontalFrameSize()).
				Align(m.columns[i].Align).
				Render(formatter.Trunc(values[i], cw))

			w = ansi.StringWidth(s)

			// Left side truncation.
			if xoffset > 0 {
				s = formatter.CutMultiline(s, xoffset+1, w)
				neww = ansi.StringWidth(s)
				xoffset -= w - neww
				w = neww
			}

			// Right side truncation.
			if w > available {
				s = formatter.Trunc(s, available)
			} else if w == available && i < len(m.columns)-1 {
				s = formatter.Trunc(s+" ", available)
			}
			available -= w

			rowOut = append(rowOut, s)

			if available <= 0 {
				break
			}
		}

		out = append(out, lipgloss.JoinHorizontal(lipgloss.Top, rowOut...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, out...)
}

func (m *Model[T]) renderScrollbar() string {
	mh := m.maxInnerTableHeight()
	return styles.Scrollbar(
		mh,
		m.TotalFilteredRows(),
		mh,
		m.yoffset,
		styles.IconScrollbar,
		styles.IconScrollbar,
	)
}

func (m *Model[T]) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	base := m.styles.Base.
		Height(m.Height).
		Width(m.Width)

	switch {
	case m.getHeaderHeight()+1 > m.Height: //nolint:gocritic
		return base.
			Align(lipgloss.Center, lipgloss.Center).
			Render("table too small to render")
	case m.loading:
		return base.Render(m.loader.View())
	case len(m.data) == 0:
		return base.Render(m.styles.NoResults.Width(m.Width).Render(m.config.NoResultsMsg))
	case len(m.filtered) == 0 && m.filter != "":
		return base.Render(
			m.styles.NoResults.
				Width(m.Width).
				Render(fmt.Sprintf(m.config.NoResultsFilterMsg, m.filter)),
		)
	}

	maxWidth := m.maxInnerTableWidth()
	if m.needsScrollbar() {
		return m.styles.Base.
			Height(m.Height).
			Width(m.Width).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					m.renderHeader(maxWidth, true),
					lipgloss.JoinHorizontal(
						lipgloss.Top,
						m.renderBody(maxWidth),
						m.renderScrollbar(),
					),
				),
			)
	}

	return m.styles.Base.
		Height(m.Height).
		Width(m.Width).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			m.renderHeader(maxWidth, false),
			m.renderBody(maxWidth),
		))
}
