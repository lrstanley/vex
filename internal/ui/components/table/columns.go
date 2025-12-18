// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// Column is a column in the table.
type Column[T Row] struct {
	ID               ID                 // Unique identifier for the column.
	Title            string             // Display name for the column.
	Disabled         bool               // Whether the column is disabled. When disabled, will not be used for filtering.
	MinWidth         int                // Optional minimum width for the column.
	MaxWidth         int                // Optional maximum width for the column.
	Align            lipgloss.Position  // Optional alignment for the column. Defaults to left.
	DisableFiltering bool               // Whether the column is disabled for filtering.
	StyleFn          CellStyleFunc[T]   // Optional function to style the column.
	AccessorFn       func(row T) string // Function which is used to access the value for the column. Must provide unless using StaticRow[[]string].
}

// validateColumns validates the columns and panics if they are invalid.
func (m *Model[T]) validateColumns() {
	if len(m.config.Columns) == 0 {
		panic("no columns provided")
	}

	var empty T

	var hasEnabledColumn bool
	for i, c := range m.config.Columns {
		if c.ID == "" {
			panic(fmt.Sprintf("ID is required for column with title %q", c.Title))
		}

		if _, ok := any(empty).(*StaticRow[[]string]); ok && c.AccessorFn == nil {
			c.AccessorFn = func(row T) string {
				return any(row).(*StaticRow[[]string]).Value[i]
			}
		}

		if c.AccessorFn == nil {
			panic(fmt.Sprintf("AccessorFn is required for column with title %q", c.Title))
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

	m.columnIDMap = make(map[ID]int, len(m.config.Columns))
	for i := range m.config.Columns {
		m.columnIDMap[m.config.Columns[i].ID] = i
	}

	m.updateCalculations()
}

// getRowValues returns the values for a row, not including disabled columns.
func (m *Model[T]) getRowValues(row T, onlyFilterable bool) []string {
	values := make([]string, len(m.config.Columns))
	for i := range m.config.Columns {
		if m.config.Columns[i].Disabled || (onlyFilterable && m.config.Columns[i].DisableFiltering) {
			continue
		}
		values[i] = m.config.Columns[i].AccessorFn(row)
	}
	return values
}

// GetColumnByID returns a column by its ID, if it exists.
func (m *Model[T]) GetColumnByID(id ID) *Column[T] {
	i, ok := m.columnIDMap[id]
	if !ok {
		return nil
	}
	return m.config.Columns[i]
}

// ToggleColumn toggles a specific column on or off.
func (m *Model[T]) ToggleColumn(id ID, enabled bool) {
	m.config.Columns[m.columnIDMap[id]].Disabled = !enabled
	m.applyFiltering()
	m.sanitizeHighlighted()
	m.updateCalculations()
}
