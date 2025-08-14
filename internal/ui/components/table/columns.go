// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import "github.com/charmbracelet/lipgloss/v2"

type Column struct { // TODO: add width.
	ID               ID
	Disabled         bool
	Title            string
	MinWidth         int
	MaxWidth         int
	Align            lipgloss.Position
	DisableFiltering bool
}

// GetColumnByID returns a column by its ID, if it exists.
func (m *Model[T]) GetColumnByID(id ID) *Column {
	i, ok := m.columnIDMap[id]
	if !ok {
		return nil
	}
	return m.columns[i]
}

// ToggleColumn toggles a specific column on or off.
func (m *Model[T]) ToggleColumn(id ID, enabled bool) {
	m.columns[m.columnIDMap[id]].Disabled = !enabled
	m.applyFiltering()
	m.sanitizeHighlighted()
	m.updateCalculations()
}

// SetColumns updates the columns, does some basic validation (at least 1 enabled,
// correct min/max widths, etc), panics if invalid, and updates internal caches.
func (m *Model[T]) SetColumns(columns []*Column) {
	m.columns = columns

	if len(m.columns) == 0 {
		panic("no columns provided")
	}

	var hasEnabledColumn bool
	for _, c := range m.columns {
		if c.ID == "" {
			panic("column ID is required")
		}
		if c.MinWidth < 0 {
			c.MinWidth = 0
		}
		if c.MaxWidth < 0 {
			c.MaxWidth = 0
		}
		if c.MaxWidth < c.MinWidth {
			c.MaxWidth = c.MinWidth
		}

		if c.Align == lipgloss.Top {
			c.Align = lipgloss.Left
		}

		if !c.Disabled {
			hasEnabledColumn = true
		}
	}

	if !hasEnabledColumn {
		panic("no enabled columns provided")
	}

	m.columnIDMap = make(map[ID]int, len(m.columns))
	for i := range m.columns {
		m.columnIDMap[m.columns[i].ID] = i
	}

	m.updateCalculations()
}
