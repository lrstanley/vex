// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/utils"
)

// getHeaderHeight returns the height of the headers.
func (m *Model[T]) getHeaderHeight() int {
	return 1 + m.styles.Header.GetVerticalFrameSize()
}

// needsScrollbar returns true if the table needs a scrollbar.
func (m *Model[T]) needsScrollbar() bool {
	if m.filter != "" {
		return m.maxInnerTableHeight() < len(m.filtered)
	}
	return m.maxInnerTableHeight() < len(m.data)
}

// maxInnerTableHeight returns the maximum height of the inner table, the free space
// left after headers, and the base style.
func (m *Model[T]) maxInnerTableHeight() int {
	return m.Height - m.getHeaderHeight() - m.styles.Base.GetVerticalFrameSize()
}

// maxInnerTableWidth returns the maximum potential width of the space inside the
// base styling.
func (m *Model[T]) maxInnerTableWidth() int {
	maxWidth := m.Width - m.styles.Base.GetHorizontalFrameSize()
	if m.needsScrollbar() {
		maxWidth--
	}
	return maxWidth
}

// innerTableWidth returns the maximum width of the inner table, not including
// the scrollbar if one is needed. This is how much space is actually consumed by
// the rendered table.
func (m *Model[T]) innerTableWidth() int {
	var totalWidth int
	for _, width := range m.maxColumnWidths {
		totalWidth += width + max(m.styles.Cell.GetHorizontalFrameSize(), m.styles.Header.GetHorizontalFrameSize())
	}
	totalWidth = min(totalWidth, m.maxInnerTableWidth())
	return totalWidth
}

// maxYOffset returns the maximum potential y offset for the table, not including
// headers.
func (m *Model[T]) maxYOffset() int {
	if m.filter != "" {
		return max(0, len(m.filtered)-m.maxInnerTableHeight())
	}
	return max(0, len(m.data)-m.maxInnerTableHeight())
}

// setYOffset sets the y offset to the given value. Note that [setSelectedIndex]
// can also udate the yoffset to ensure the selected index is within the view.
func (m *Model[T]) setYOffset(y int) {
	m.yoffset = utils.Clamp(y, 0, m.maxYOffset())

	if m.yoffset > m.selectedIndex {
		// Selected index is above the view.
		m.selectedIndex = utils.Clamp(m.yoffset, 0, m.TotalFilteredRows()-1)
	} else if bounds := m.yoffset + m.maxInnerTableHeight(); m.selectedIndex >= bounds {
		// Selected index is below the view.
		m.selectedIndex = utils.Clamp(bounds-1, 0, m.TotalFilteredRows()-1)
	}
}

// setSelectedIndex moves the selected index (row). Note that [setYOffset] can also
// update the selected index to ensure it is within the view.
func (m *Model[T]) setSelectedIndex(i int) {
	m.selectedIndex = utils.Clamp(i, 0, m.TotalFilteredRows()-1)

	if m.selectedIndex < m.yoffset {
		// Selected index is above the view, so move the view up to the selected index.
		m.yoffset = utils.Clamp(m.selectedIndex, 0, m.maxYOffset())
	} else if bounds := m.yoffset + m.maxInnerTableHeight(); m.selectedIndex >= bounds {
		// Selected index is below the view, so move the view down to the selected index.
		m.yoffset = utils.Clamp(m.selectedIndex-m.maxInnerTableHeight()+1, 0, m.maxYOffset())
	}
}

// maxXOffset returns the maximum potential x offset for the table.
func (m *Model[T]) maxXOffset() int {
	maxWidth := m.maxInnerTableWidth()
	var width int
	for _, w := range m.maxColumnWidths {
		width += w + max(m.styles.Cell.GetHorizontalFrameSize(), m.styles.Header.GetHorizontalFrameSize())
	}
	if width <= maxWidth {
		return 0
	}
	return max(0, width-maxWidth)
}

// setXOffset sets the x offset to the given value.
func (m *Model[T]) setXOffset(x int) {
	m.xoffset = utils.Clamp(x, 0, m.maxXOffset())
}

// SetDimensions sets the dimensions of the table. Prefer this over [SetWidth] and
// [SetHeight] as it will be more efficient.
func (m *Model[T]) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
	m.setSelectedIndex(m.selectedIndex)
	m.setXOffset(m.xoffset)
	m.setYOffset(m.yoffset)
	m.updateCalculations()
}

// updateCalculations updates the internal calculations for the table, like column
// min/max widths, and handle padding to fill the table.
func (m *Model[T]) updateCalculations() {
	var w int
	var values []string
	var lastEnabledColumn int

	// Pre-seed the column width calculations using the title of the column.
	for i := range m.config.Columns {
		if m.config.Columns[i].Disabled {
			delete(m.maxColumnWidths, m.config.Columns[i].ID)
			continue
		}
		w = ansi.StringWidth(m.config.Columns[i].Title)
		if mw := m.config.Columns[i].MinWidth; mw > 0 && w < mw {
			w = mw
		}
		if mw := m.config.Columns[i].MaxWidth; mw > 0 && w > mw {
			w = mw
		}
		m.maxColumnWidths[m.config.Columns[i].ID] = w
		lastEnabledColumn = i
	}

	// If there is a row cell that is larger than the pre-seeded width, then
	// update the column width to the larger value.
	for row := range m.GetRows() {
		values = m.getRowValues(row, false)
		for i := range m.config.Columns {
			if m.config.Columns[i].Disabled {
				continue
			}
			w = ansi.StringWidth(values[i])
			if mw := m.config.Columns[i].MaxWidth; mw > 0 && w > mw {
				w = mw
			}
			m.maxColumnWidths[m.config.Columns[i].ID] = max(
				w,
				m.maxColumnWidths[m.config.Columns[i].ID],
			)
		}
	}

	// if the total inner width is less than the max potential inner width, add
	// additional spacing to the last column.
	available := m.maxInnerTableWidth() - m.innerTableWidth()
	if available > 0 {
		m.maxColumnWidths[m.config.Columns[lastEnabledColumn].ID] += available
	}
}

// SetWidth sets the total width of the table.
func (m *Model[T]) SetWidth(width int) {
	m.SetDimensions(width, m.Height)
}

// SetHeight sets the total height of the table.
func (m *Model[T]) SetHeight(height int) {
	m.SetDimensions(m.Width, height)
}

// MoveUp moves the selected row up by the given number of rows.
func (m *Model[T]) MoveUp(n int) {
	m.setSelectedIndex(m.selectedIndex - n)
}

// MoveDown moves the selected row down by the given number of rows.
func (m *Model[T]) MoveDown(n int) {
	m.setSelectedIndex(m.selectedIndex + n)
}

// MoveLeft moves the selected row left by the given number of columns.
func (m *Model[T]) MoveLeft(n int) {
	m.setXOffset(m.xoffset - n)
}

// MoveRight moves the selected row right by the given number of columns.
func (m *Model[T]) MoveRight(n int) {
	m.setXOffset(m.xoffset + n)
}

// GoToTop moves the selected row, and view, to the top of the table.
func (m *Model[T]) GoToTop() {
	m.setSelectedIndex(0)
}

// GoToBottom moves the selected row, and view, to the bottom of the table.
func (m *Model[T]) GoToBottom() {
	if m.filter != "" {
		m.setSelectedIndex(len(m.filtered) - 1)
	} else {
		m.setSelectedIndex(len(m.data) - 1)
	}
}
