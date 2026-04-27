// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/steep"
)

// testRow implements the Row interface for testing.
type testRow struct {
	id   string
	data []string
}

func mutate(tb testing.TB, tm *steep.Harness, fn func()) {
	tb.Helper()

	steep.Mutate(tb, tm, func(m *Model[testRow]) *Model[testRow] {
		fn()
		return m
	})
	tm.WaitSettleMessages(tb, steep.WithSettleTimeout(25*time.Millisecond), steep.WithCheckInterval(5*time.Millisecond))
}

func (r testRow) ID() ID {
	return ID(r.id)
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("basic-table", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "description", Title: "Description", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(30, 5))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"item1", "description1"}},
				{id: "2", data: []string{"item2", "description2"}},
			})
		})

		tm.WaitContainsStrings(t, []string{"item1", "item2", "Name", "Description"})
		tm.RequireDimensions(t, m.GetWidth(), m.GetHeight()).
			RequireSnapshotNoANSI(t)
	})

	t.Run("loading-state", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
		}

		m := New(app, Config[testRow]{Columns: columns})
		m.loading = true

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(40, 5))
		tm.WaitContainsString(t, "loading")
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("no-results", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
		}

		m := New(app, Config[testRow]{
			Columns:      columns,
			NoResultsMsg: "no items found",
		})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(40, 5))
		mutate(t, tm, func() {
			m.SetRows([]testRow{})
		})
		tm.WaitContainsString(t, "no items found")
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("filtered-no-results", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
		}

		m := New(app, Config[testRow]{
			Columns:            columns,
			NoResultsFilterMsg: "no results for %q",
		})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(60, 5))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"item1"}},
				{id: "2", data: []string{"item2"}},
			})
			m.SetFilter("nonexistent")
		})

		tm.WaitContainsString(t, "no results for")
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(0, 0))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"item1"}},
			})
		})
		tm.RequireStringNotContains(t, "item1").WaitSettleView(t).
			RequireDimensions(t, 0, 0)
	})
}

func TestTableScrolling(t *testing.T) {
	t.Parallel()

	t.Run("vertical-scrollbar", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "value", Title: "Value", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(30, steep.DefaultTermHeight))

		// Create many rows to exceed the height
		var rows []testRow
		for i := range steep.DefaultTermHeight * 2 {
			rows = append(rows, testRow{
				id:   string(rune('a' + i)),
				data: []string{fmt.Sprintf("item%d", i), fmt.Sprintf("value%d", i)},
			})
		}
		mutate(t, tm, func() {
			m.SetRows(rows)
		})

		tm.WaitContainsStrings(t, []string{"item0", "item1", "item2"})
		tm.RequireSnapshotNoANSI(t)

		mutate(t, tm, func() {
			m.MoveDown(steep.DefaultTermHeight - 1 + 5)
		})
		tm.WaitContainsString(t, fmt.Sprintf("item%d", len(slices.Collect(m.GetVisibleRows()))+5))
		tm.RequireSnapshotNoANSI(t)

		mutate(t, tm, func() {
			m.GoToBottom()
		})
		tm.WaitContainsString(t, fmt.Sprintf("item%d", len(rows)-1))
		tm.RequireSnapshotNoANSI(t)

		mutate(t, tm, func() {
			m.GoToTop()
		})
		tm.WaitContainsString(t, "item0")
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("horizontal-truncation", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "description", Title: "Very Long Description Column", AccessorFn: func(row testRow) string { return row.data[1] }},
			{ID: "value", Title: "Value", AccessorFn: func(row testRow) string { return row.data[2] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		// Set small width to force horizontal scrolling
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(55, 5))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"item1", "This is a very long description that will cause horizontal scrolling", "value1"}},
				{id: "2", data: []string{"item2", "Another long description for testing horizontal scroll", "value2"}},
			})
		})

		tm.WaitContainsString(t, "This is a very long description")
		tm.RequireSnapshotNoANSI(t)

		// As far right as possible.
		mutate(t, tm, func() {
			m.MoveRight(100)
		})
		tm.WaitContainsString(t, "value1")
		tm.RequireSnapshotNoANSI(t)

		// As far left as possible.
		mutate(t, tm, func() {
			m.MoveLeft(100)
		})
		tm.WaitContainsString(t, "This is a very long description")
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("too-small-x-y", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "description", Title: "Very Long Description Column", AccessorFn: func(row testRow) string { return row.data[1] }},
			{ID: "value", Title: "Value", AccessorFn: func(row testRow) string { return row.data[2] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		// Set small dimensions to force both scrollbars
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(55, 10))

		// Create many rows with long content
		var rows []testRow
		for i := range steep.DefaultTermHeight * 2 {
			rows = append(rows, testRow{
				id:   string(rune('a' + i)),
				data: []string{fmt.Sprintf("item%d", i), fmt.Sprintf("This is a very long description for item %d that will cause horizontal scrolling", i), fmt.Sprintf("value%d", i)},
			})
		}
		mutate(t, tm, func() {
			m.SetRows(rows)
		})
		tm.WaitContainsString(t, "This is a very long")
		tm.RequireSnapshotNoANSI(t)

		// Test scrolling both directions
		mutate(t, tm, func() {
			m.MoveDown(5)
			m.MoveRight(15)
		})
		tm.RequireSnapshotNoANSI(t)
	})
}

func TestTableTruncation(t *testing.T) {
	t.Parallel()

	t.Run("column-truncation", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", MaxWidth: 10, AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "description", Title: "Description", MaxWidth: 15, AccessorFn: func(row testRow) string { return row.data[1] }},
			{ID: "value", Title: "Value", MaxWidth: 8, AccessorFn: func(row testRow) string { return row.data[2] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(60, 5))

		rows := []testRow{
			{id: "1", data: []string{"Very Long Name That Should Be Truncated", "This is a very long description that should be truncated", "123456789"}},
			{id: "2", data: []string{"Another Long Name", "Another long description", "987654321"}},
		}
		mutate(t, tm, func() {
			m.SetRows(rows)
		})

		tm.WaitContainsString(t, "Very Long")
		tm.RequireStringNotContains(t, "Very Long Name").
			RequireSnapshotNoANSI(t)
	})

	t.Run("ellipsis-indicator", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "description", Title: "Description", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(20, 4))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"Very Long Name", "Very Long Description"}},
			})
		})

		tm.WaitContainsString(t, "Very Long Name")
		tm.RequireSnapshotNoANSI(t)
	})
}

func TestTableFiltering(t *testing.T) {
	t.Parallel()

	t.Run("basic-filtering", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "category", Title: "Category", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(30, 6))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"apple", "fruit"}},
				{id: "2", data: []string{"banana", "fruit"}},
				{id: "3", data: []string{"carrot", "vegetable"}},
				{id: "4", data: []string{"broccoli", "vegetable"}},
			})
		})

		// Test filtering by name
		mutate(t, tm, func() {
			m.SetFilter("apple")
		})
		tm.WaitContainsString(t, "apple")
		tm.RequireStringNotContains(t, "banana", "carrot", "broccoli").
			RequireSnapshotNoANSI(t)

		// Test filtering by category
		mutate(t, tm, func() {
			m.SetFilter("fruit")
		})
		tm.WaitContainsStrings(t, []string{"apple", "banana"})
		tm.RequireStringNotContains(t, "carrot", "broccoli").
			RequireSnapshotNoANSI(t)

		// Test clearing filter
		mutate(t, tm, func() {
			m.SetFilter("")
		})
		tm.WaitContainsStrings(t, []string{"apple", "banana", "carrot", "broccoli"})
		tm.RequireSnapshotNoANSI(t)
	})
}

func TestTableSelection(t *testing.T) {
	t.Parallel()

	t.Run("row-selection", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "value", Title: "Value", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(30, 5))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"item1", "value1"}},
				{id: "2", data: []string{"item2", "value2"}},
				{id: "3", data: []string{"item3", "value3"}},
			})
		})

		tm.WaitContainsStrings(t, []string{"item1", "item2", "item3"})

		// Test initial selection
		selected, found := m.GetSelectedRow()
		if !found {
			t.Fatal("expected to have a selected row")
		}
		if selected.ID() != "1" {
			t.Fatalf("expected first row to be selected, got %s", selected.ID())
		}

		// Test moving selection
		mutate(t, tm, func() {
			m.MoveDown(1)
		})
		selected, found = m.GetSelectedRow()
		if !found {
			t.Fatal("expected to have a selected row")
		}
		if selected.ID() != "2" {
			t.Fatalf("expected second row to be selected, got %s", selected.ID())
		}

		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("selection-with-filter", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "category", Title: "Category", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(30, 5))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"apple", "fruit"}},
				{id: "2", data: []string{"banana", "fruit"}},
				{id: "3", data: []string{"carrot", "vegetable"}},
			})
		})

		tm.WaitContainsStrings(t, []string{"apple", "banana", "carrot"})

		// Select second row
		mutate(t, tm, func() {
			m.MoveDown(1)
		})
		tm.WaitContainsStrings(t, []string{"apple", "banana", "carrot"})

		// Apply filter
		mutate(t, tm, func() {
			m.SetFilter("fruit")
		})
		tm.WaitContainsStrings(t, []string{"apple", "banana"})
		tm.RequireStringNotContains(t, "carrot")

		// Check that selection is maintained
		selected, found := m.GetSelectedRow()
		if !found {
			t.Fatal("expected to have a selected row")
		}
		if selected.ID() != "2" {
			t.Fatalf("expected second row to be selected, got %s", selected.ID())
		}

		tm.RequireSnapshotNoANSI(t)
	})
}

func TestTableColumns(t *testing.T) {
	t.Parallel()

	t.Run("column-management", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "description", Title: "Description", AccessorFn: func(row testRow) string { return row.data[1] }},
			{ID: "value", Title: "Value", AccessorFn: func(row testRow) string { return row.data[2] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(30, 5))

		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"item1", "description1", "value1"}},
				{id: "2", data: []string{"item2", "description2", "value2"}},
			})
		})

		// Test disabling a column
		mutate(t, tm, func() {
			m.ToggleColumn("description", false)
		})
		tm.WaitContainsStrings(t, []string{"item1", "item2"})
		tm.RequireStringNotContains(t, "description1", "description2").
			RequireSnapshotNoANSI(t)

		// Test re-enabling a column
		mutate(t, tm, func() {
			m.ToggleColumn("description", true)
		})
		tm.WaitContainsStrings(t, []string{"item1", "item2"})
		tm.WaitContainsStrings(t, []string{"description1", "description2"})
		tm.RequireSnapshotNoANSI(t)

		// Test getting column by ID
		col := m.GetColumnByID("name")
		if col == nil {
			t.Fatal("expected to find column 'name'")
		}
		if col.Title != "Name" {
			t.Fatalf("expected column title 'Name', got '%s'", col.Title)
		}
	})
}

func TestTableData(t *testing.T) {
	t.Parallel()

	t.Run("data-operations", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "value", Title: "Value", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(60, 8))

		// Test initial empty state
		if m.TotalRows() != 0 {
			t.Fatalf("expected 0 rows initially, got %d", m.TotalRows())
		}

		// Test adding rows
		mutate(t, tm, func() {
			m.SetRows([]testRow{
				{id: "1", data: []string{"item1", "value1"}},
				{id: "2", data: []string{"item2", "value2"}},
			})
		})

		if m.TotalRows() != 2 {
			t.Fatalf("expected 2 rows, got %d", m.TotalRows())
		}

		// Test updating a row
		updatedRow := testRow{id: "1", data: []string{"updated_item1", "updated_value1"}}
		mutate(t, tm, func() {
			m.UpdateRow(updatedRow)
		})

		// Test getting row by ID
		retrievedRow := m.GetRowByID("1")
		if retrievedRow.data[0] != "updated_item1" {
			t.Fatalf("expected updated row, got %v", retrievedRow)
		}

		// Test prepending a row
		newRow := testRow{id: "0", data: []string{"item0", "value0"}}
		mutate(t, tm, func() {
			m.PrependRow(newRow)
		})

		if m.TotalRows() != 3 {
			t.Fatalf("expected 3 rows after prepend, got %d", m.TotalRows())
		}

		// Test appending a row
		appendRow := testRow{id: "3", data: []string{"item3", "value3"}}
		mutate(t, tm, func() {
			m.AppendRow(appendRow)
		})

		if m.TotalRows() != 4 {
			t.Fatalf("expected 4 rows after append, got %d", m.TotalRows())
		}

		// Test deleting a row
		mutate(t, tm, func() {
			m.DeleteRowByID("2")
		})

		if m.TotalRows() != 3 {
			t.Fatalf("expected 3 rows after delete, got %d", m.TotalRows())
		}

		tm.RequireSnapshotNoANSI(t)
	})
}

func TestTableStyles(t *testing.T) {
	t.Parallel()

	t.Run("style-customization", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "value", Title: "Value", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(30, 5))

		rows := []testRow{
			{id: "1", data: []string{"item1", "value1"}},
			{id: "2", data: []string{"item2", "value2"}},
		}
		mutate(t, tm, func() {
			m.SetRows(rows)
		})

		customStyles := Styles{
			Header: lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false),
		}
		mutate(t, tm, func() {
			m.SetStyles(customStyles)
		})

		tm.RequireDimensions(t, m.GetWidth(), m.GetHeight()).
			RequireSnapshotNoANSI(t)
	})
}
