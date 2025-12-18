// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/testui"
)

// testRow implements the Row interface for testing.
type testRow struct {
	id   string
	data []string
}

func (r testRow) ID() ID {
	return ID(r.id)
}

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true}) //nolint:errcheck
	os.Exit(v)
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("basic-table", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name"},
			{ID: "description", Title: "Description"},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(30, 5))

		m.SetRows([]testRow{
			{id: "1", data: []string{"item1", "description1"}},
			{id: "2", data: []string{"item2", "description2"}},
		})

		tm.ExpectViewContains(t, "item1", "item2", "Name", "Description")
		tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
		tm.ExpectViewSnapshot(t)
	})

	t.Run("loading-state", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
		}

		m := New(app, Config[testRow]{Columns: columns})
		m.loading = true

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(40, 5))
		tm.ExpectViewContains(t, "loading")
		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(40, 5))
		m.SetRows([]testRow{})
		tm.ExpectViewContains(t, "no items found")
		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(60, 5))

		m.SetRows([]testRow{
			{id: "1", data: []string{"item1"}},
			{id: "2", data: []string{"item2"}},
		})
		m.SetFilter("nonexistent")

		tm.ExpectViewContains(t, "no results for")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(0, 0))

		m.SetRows([]testRow{
			{id: "1", data: []string{"item1"}},
		})
		time.Sleep(100 * time.Millisecond)
		tm.ExpectViewNotContains(t, "item1")
		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(30, testui.DefaultTermHeight))

		// Create many rows to exceed the height
		var rows []testRow
		for i := range testui.DefaultTermHeight * 2 {
			rows = append(rows, testRow{
				id:   string(rune('a' + i)),
				data: []string{fmt.Sprintf("item%d", i), fmt.Sprintf("value%d", i)},
			})
		}
		m.SetRows(rows)

		tm.ExpectViewContains(t, "item0", "item1", "item2")
		tm.ExpectViewSnapshot(t)

		m.MoveDown(testui.DefaultTermHeight - 1 + 5)
		tm.ExpectViewContains(t, fmt.Sprintf("item%d", len(slices.Collect(m.GetVisibleRows()))+5))
		tm.ExpectViewSnapshot(t)

		m.GoToBottom()
		tm.ExpectViewContains(t, fmt.Sprintf("item%d", len(rows)-1))
		tm.ExpectViewSnapshot(t)

		m.GoToTop()
		tm.ExpectViewContains(t, "item0")
		tm.ExpectViewSnapshot(t)
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
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(55, 5))

		m.SetRows([]testRow{
			{id: "1", data: []string{"item1", "This is a very long description that will cause horizontal scrolling", "value1"}},
			{id: "2", data: []string{"item2", "Another long description for testing horizontal scroll", "value2"}},
		})

		tm.ExpectViewContains(t, "This is a very long description")
		tm.ExpectViewSnapshot(t)

		// As far right as possible.
		m.MoveRight(100)
		tm.ExpectViewContains(t, "value1")
		tm.ExpectViewSnapshot(t)

		// As far left as possible.
		m.MoveLeft(100)
		tm.ExpectViewContains(t, "This is a very long description")
		tm.ExpectViewSnapshot(t)
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
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(55, 10))

		// Create many rows with long content
		var rows []testRow
		for i := range testui.DefaultTermHeight * 2 {
			rows = append(rows, testRow{
				id:   string(rune('a' + i)),
				data: []string{fmt.Sprintf("item%d", i), fmt.Sprintf("This is a very long description for item %d that will cause horizontal scrolling", i), fmt.Sprintf("value%d", i)},
			})
		}
		m.SetRows(rows)
		tm.ExpectViewContains(t, "This is a very long")
		tm.ExpectViewSnapshot(t)

		// Test scrolling both directions
		m.MoveDown(5)
		m.MoveRight(15)
		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(60, 5))

		rows := []testRow{
			{id: "1", data: []string{"Very Long Name That Should Be Truncated", "This is a very long description that should be truncated", "123456789"}},
			{id: "2", data: []string{"Another Long Name", "Another long description", "987654321"}},
		}
		m.SetRows(rows)

		tm.ExpectViewContains(t, "Very Long")
		tm.ExpectViewNotContains(t, "Very Long Name")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("ellipsis-indicator", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "description", Title: "Description", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(20, 4))

		m.SetRows([]testRow{
			{id: "1", data: []string{"Very Long Name", "Very Long Description"}},
		})

		tm.ExpectViewContains(t, "Very Long Name")
		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(30, 6))

		m.SetRows([]testRow{
			{id: "1", data: []string{"apple", "fruit"}},
			{id: "2", data: []string{"banana", "fruit"}},
			{id: "3", data: []string{"carrot", "vegetable"}},
			{id: "4", data: []string{"broccoli", "vegetable"}},
		})

		// Test filtering by name
		m.SetFilter("apple")
		tm.ExpectViewContains(t, "apple")
		tm.ExpectViewNotContains(t, "banana", "carrot", "broccoli")
		tm.ExpectViewSnapshot(t)

		// Test filtering by category
		m.SetFilter("fruit")
		tm.ExpectViewContains(t, "apple", "banana")
		tm.ExpectViewNotContains(t, "carrot", "broccoli")
		tm.ExpectViewSnapshot(t)

		// Test clearing filter
		m.SetFilter("")
		tm.ExpectViewContains(t, "apple", "banana", "carrot", "broccoli")
		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(30, 5))

		m.SetRows([]testRow{
			{id: "1", data: []string{"item1", "value1"}},
			{id: "2", data: []string{"item2", "value2"}},
			{id: "3", data: []string{"item3", "value3"}},
		})

		tm.ExpectViewContains(t, "item1", "item2", "item3")

		// Test initial selection
		selected, found := m.GetSelectedRow()
		if !found {
			t.Fatal("expected to have a selected row")
		}
		if selected.ID() != "1" {
			t.Fatalf("expected first row to be selected, got %s", selected.ID())
		}

		// Test moving selection
		m.MoveDown(1)
		selected, found = m.GetSelectedRow()
		if !found {
			t.Fatal("expected to have a selected row")
		}
		if selected.ID() != "2" {
			t.Fatalf("expected second row to be selected, got %s", selected.ID())
		}

		tm.ExpectViewSnapshot(t)
	})

	t.Run("selection-with-filter", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*Column[testRow]{
			{ID: "name", Title: "Name", AccessorFn: func(row testRow) string { return row.data[0] }},
			{ID: "category", Title: "Category", AccessorFn: func(row testRow) string { return row.data[1] }},
		}

		m := New(app, Config[testRow]{Columns: columns})

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(30, 5))

		m.SetRows([]testRow{
			{id: "1", data: []string{"apple", "fruit"}},
			{id: "2", data: []string{"banana", "fruit"}},
			{id: "3", data: []string{"carrot", "vegetable"}},
		})

		tm.ExpectViewContains(t, "apple", "banana", "carrot")

		// Select second row
		m.MoveDown(1)
		tm.ExpectViewContains(t, "apple", "banana", "carrot")

		// Apply filter
		m.SetFilter("fruit")
		tm.ExpectViewContains(t, "apple", "banana")
		tm.ExpectViewNotContains(t, "carrot")

		// Check that selection is maintained
		selected, found := m.GetSelectedRow()
		if !found {
			t.Fatal("expected to have a selected row")
		}
		if selected.ID() != "2" {
			t.Fatalf("expected second row to be selected, got %s", selected.ID())
		}

		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(30, 5))

		m.SetRows([]testRow{
			{id: "1", data: []string{"item1", "description1", "value1"}},
			{id: "2", data: []string{"item2", "description2", "value2"}},
		})

		// Test disabling a column
		m.ToggleColumn("description", false)
		tm.ExpectViewContains(t, "item1", "item2")
		tm.ExpectViewNotContains(t, "description1", "description2")
		tm.ExpectViewSnapshot(t)

		// Test re-enabling a column
		m.ToggleColumn("description", true)
		tm.ExpectViewContains(t, "item1", "item2")
		tm.ExpectViewContains(t, "description1", "description2")
		tm.ExpectViewSnapshot(t)

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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(60, 8))

		// Test initial empty state
		if m.TotalRows() != 0 {
			t.Fatalf("expected 0 rows initially, got %d", m.TotalRows())
		}

		// Test adding rows
		m.SetRows([]testRow{
			{id: "1", data: []string{"item1", "value1"}},
			{id: "2", data: []string{"item2", "value2"}},
		})

		if m.TotalRows() != 2 {
			t.Fatalf("expected 2 rows, got %d", m.TotalRows())
		}

		// Test updating a row
		updatedRow := testRow{id: "1", data: []string{"updated_item1", "updated_value1"}}
		m.UpdateRow(updatedRow)

		// Test getting row by ID
		retrievedRow := m.GetRowByID("1")
		if retrievedRow.data[0] != "updated_item1" {
			t.Fatalf("expected updated row, got %v", retrievedRow)
		}

		// Test prepending a row
		newRow := testRow{id: "0", data: []string{"item0", "value0"}}
		m.PrependRow(newRow)

		if m.TotalRows() != 3 {
			t.Fatalf("expected 3 rows after prepend, got %d", m.TotalRows())
		}

		// Test appending a row
		appendRow := testRow{id: "3", data: []string{"item3", "value3"}}
		m.AppendRow(appendRow)

		if m.TotalRows() != 4 {
			t.Fatalf("expected 4 rows after append, got %d", m.TotalRows())
		}

		// Test deleting a row
		m.DeleteRowByID("2")

		if m.TotalRows() != 3 {
			t.Fatalf("expected 3 rows after delete, got %d", m.TotalRows())
		}

		tm.ExpectViewSnapshot(t)
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(30, 5))

		rows := []testRow{
			{id: "1", data: []string{"item1", "value1"}},
			{id: "2", data: []string{"item2", "value2"}},
		}
		m.SetRows(rows)

		customStyles := Styles{
			Header: lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false),
		}
		m.SetStyles(customStyles)

		tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
		tm.ExpectViewSnapshot(t)
	})
}
