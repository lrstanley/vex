// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"iter"
	"slices"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/fuzzy"
)

// StaticRow is a row that is created from a static value and ID.
type StaticRow[T any] struct {
	Value   T
	ValueID ID
}

func (r *StaticRow[T]) ID() ID {
	return r.ValueID
}

// RowsFrom creates a slice of rows from a slice of data using a function to
// get the ID of each row, which satisfies the [Row] interface.
func RowsFrom[T any](data []T, idFn func(T) ID) []*StaticRow[T] {
	out := make([]*StaticRow[T], len(data))
	for i := range data {
		out[i] = &StaticRow[T]{
			Value:   data[i],
			ValueID: idFn(data[i]),
		}
	}
	return out
}

type Row interface {
	ID() ID
}

// applyFiltering applies the current filter to the data.
func (m *Model[T]) applyFiltering() {
	if m.filter == "" {
		m.filtered = nil
		return
	}

	ids := m.AllIDs()
	var values []string
	filterable := make([]string, 0, len(m.columns))

	m.filtered = fuzzy.FindRankedRow(m.filter, ids, func(id ID) []string {
		i := m.dataIDMap[id]
		values = m.config.RowFn(m.data[i])
		filterable = filterable[:0]
		for i := range m.columns {
			if m.columns[i].DisableFiltering {
				continue
			}
			filterable = append(filterable, values[i])
		}
		return filterable
	})
}

// sanitizeHighlighted sanitizes the stored highlighted rows to only include rows
// that are in the filtered or unfiltered data.
func (m *Model[T]) sanitizeHighlighted() {
	if !m.config.AllowHighlighting {
		m.highlighted = nil
		return
	}

	if m.filter != "" {
		m.highlighted = slices.DeleteFunc(m.highlighted, func(id ID) bool {
			return !slices.Contains(m.filtered, id)
		})
	} else {
		ids := m.IDs()
		m.highlighted = slices.DeleteFunc(m.highlighted, func(id ID) bool {
			return !slices.Contains(ids, id)
		})
	}
}

// GetRowByID returns a row by its ID, if it exists.
func (m *Model[T]) GetRowByID(id ID) T {
	var v T
	i, ok := m.dataIDMap[id]
	if !ok {
		return v
	}
	return m.data[i]
}

// SetSelected sets the selected row to the given ID, if it exists. If it doesn't,
// it will default to the first row.
func (m *Model[T]) SetSelected(id ID) {
	if m.filter != "" {
		i := slices.Index(m.filtered, id)
		if i != -1 {
			m.setSelectedIndex(i)
			return
		}
	} else {
		if i, ok := m.dataIDMap[id]; ok {
			m.setSelectedIndex(i)
			return
		}
	}

	// Default to the first row.
	m.setSelectedIndex(0)
}

// SetRows sets the data for the table.
func (m *Model[T]) SetRows(values []T) {
	m.data = values
	m.updateDataIDMap()
	m.applyFiltering()
	m.sanitizeHighlighted()
	m.updateCalculations()
	m.setYOffset(m.yoffset)
	m.setXOffset(m.xoffset)
	m.loading = false
}

// updateDataIDMap updates the dataIDMap to reflect the current data. This should
// ALWAYS be called right after data has been added, removed, etc.
func (m *Model[T]) updateDataIDMap() {
	m.dataIDMap = make(map[ID]int, len(m.data))
	for i := range m.data {
		m.dataIDMap[m.data[i].ID()] = i
	}
}

// UpdateRow updates a row in the table, if it exists.
func (m *Model[T]) UpdateRow(row T) {
	for i := range m.data {
		if m.data[i].ID() != row.ID() {
			continue
		}
		m.data[i] = row
		m.applyFiltering()
		m.sanitizeHighlighted()
		m.updateCalculations()
		m.setYOffset(m.yoffset)
		m.setXOffset(m.xoffset)
		break
	}
}

// DeleteRowByID deletes a row by its ID, if it exists.
func (m *Model[T]) DeleteRowByID(id ID) {
	lenBefore := len(m.data)
	m.data = slices.Clip(slices.DeleteFunc(m.data, func(row T) bool {
		return row.ID() == id
	}))
	if lenBefore != len(m.data) {
		m.updateDataIDMap()
		m.applyFiltering()
		m.sanitizeHighlighted()
		m.updateCalculations()
		m.setYOffset(m.yoffset)
		m.setXOffset(m.xoffset)
	}
}

// PrependRow prepends a row to the table.
func (m *Model[T]) PrependRow(row T) {
	ids := m.IDs()
	if slices.Contains(ids, row.ID()) {
		return
	}
	m.data = append([]T{row}, m.data...)
	m.updateDataIDMap()
	m.applyFiltering()
	m.sanitizeHighlighted()
	m.updateCalculations()
	m.setYOffset(m.yoffset)
	m.setXOffset(m.xoffset)
}

// AppendRow appends a row to the table.
func (m *Model[T]) AppendRow(row T) {
	ids := m.IDs()
	if slices.Contains(ids, row.ID()) {
		return
	}
	m.data = append(m.data, row)
	m.updateDataIDMap()
	m.applyFiltering()
	m.sanitizeHighlighted()
	m.updateCalculations()
	m.setYOffset(m.yoffset)
	m.setXOffset(m.xoffset)
}

// Fetch fetches the data from the [Config.FetchFn].
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

// GetAllRows returns the data.
func (m *Model[T]) GetAllRows() []T {
	return m.data
}

// IDs returns the IDs of the data. If filtering is enabled, it will return the
// filtered IDs. See also [Model.AllIDs].
func (m *Model[T]) IDs() []ID {
	var ids []ID
	if m.filter != "" {
		return m.filtered
	}
	for i := range m.data {
		ids = append(ids, m.data[i].ID())
	}
	return ids
}

// AllIDs returns all of the IDs of the data. See also [Model.IDs].
func (m *Model[T]) AllIDs() []ID {
	ids := make([]ID, len(m.data))
	for i := range m.data {
		ids[i] = m.data[i].ID()
	}
	return ids
}

// GetRows returns an iterator of rows. If filtering is enabled, it will
// return the filtered results.
func (m *Model[T]) GetRows() iter.Seq[T] {
	return func(yield func(T) bool) {
		if m.filter != "" {
			for _, id := range m.filtered {
				if !yield(m.GetRowByID(id)) {
					return
				}
			}
			return
		}
		for _, v := range m.data {
			if !yield(v) {
				return
			}
		}
	}
}

// GetVisibleRows returns an iterator of rows that are visible in the current view.
func (m *Model[T]) GetVisibleRows() iter.Seq[T] {
	return func(yield func(T) bool) {
		if m.filter != "" {
			for _, id := range m.filtered[m.yoffset:min(m.yoffset+m.maxInnerTableHeight(), len(m.filtered))] {
				if !yield(m.GetRowByID(id)) {
					return
				}
			}
			return
		}
		for _, v := range m.data[m.yoffset:min(m.yoffset+m.maxInnerTableHeight(), len(m.data))] {
			if !yield(v) {
				return
			}
		}
	}
}

// GetFilteredRows returns an iterator of rows that are filtered.
func (m *Model[T]) GetFilteredRows() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, id := range m.filtered {
			if !yield(m.GetRowByID(id)) {
				return
			}
		}
	}
}

// GetHighlightedRows returns an iterator of rows that are highlighted.
func (m *Model[T]) GetHighlightedRows() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, id := range m.highlighted {
			if !yield(m.GetRowByID(id)) {
				return
			}
		}
	}
}

// GetSelectedRow returns the currently selected row. If there is no selected row
// (e.g. no data), will return false.
func (m *Model[T]) GetSelectedRow() (v T, found bool) {
	if m.filter != "" {
		if len(m.filtered) == 0 || m.selectedIndex >= len(m.filtered) {
			return v, false
		}
		return m.GetRowByID(m.filtered[m.selectedIndex]), true
	}
	if len(m.data) == 0 || m.selectedIndex >= len(m.data) {
		return v, false
	}
	return m.data[m.selectedIndex], true
}

// TotalRows returns the total number of data entries.
func (m *Model[T]) TotalRows() int {
	return len(m.data)
}

// TotalFilteredRows returns the number of filtered rows. If filtering is not
// enabled, it will return the total number of rows ignoring filtering (e.g.
// [Model.TotalRows]).
func (m *Model[T]) TotalFilteredRows() int {
	if m.filter != "" {
		return len(m.filtered)
	}
	return len(m.data)
}
