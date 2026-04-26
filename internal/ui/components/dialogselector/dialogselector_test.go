// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package dialogselector

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/steep"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-selector", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		m := New(app, Config{
			Columns: []*table.Column[*table.StaticRow[[]string]]{
				{ID: "name", Title: "Name"},
				{ID: "description", Title: "Description"},
			},
			FilterPlaceholder: "type to filter",
			SelectFunc:        func(_ string) tea.Cmd { return nil },
		})

		m.SetItems([][]string{
			{"item1", "item1", "description 1"},
			{"item2", "item2", "description 2"},
			{"item3", "item3", "description 3"},
		})

		tm := steep.NewViewModel(t, m)
		tm.WaitContainsString(t, "description 1")
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)
	})

	t.Run("empty-list", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		m := New(app, Config{
			Columns: []*table.Column[*table.StaticRow[[]string]]{
				{ID: "name", Title: "Name"},
				{ID: "description", Title: "Description"},
			},
			SelectFunc: func(_ string) tea.Cmd { return nil },
		})

		m.SetItems([][]string{})

		tm := steep.NewViewModel(t, m)
		tm.WaitContainsString(t, "no results found")
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		m := New(app, Config{
			Columns: []*table.Column[*table.StaticRow[[]string]]{
				{ID: "name", Title: "Name"},
				{ID: "description", Title: "Description"},
			},
			SelectFunc: func(_ string) tea.Cmd { return nil },
		})

		// Set items after creation
		// First element is the ID, subsequent elements are the column data
		m.SetItems([][]string{{"item1", "item1", "description"}})

		steep.NewViewModel(t, m, steep.WithInitialTermSize(0, 0)).
			WaitSettleView(t).
			ExpectDimensions(t, 0, 0)
	})
}

func TestDialogSelectorFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("filtering", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column[*table.StaticRow[[]string]]{
			{ID: "name", Title: "Name"},
			{ID: "category", Title: "Category"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(_ string) tea.Cmd {
				return nil
			},
		})

		// Set items with different categories
		// Note: First element is the ID, subsequent elements are column data
		m.SetItems([][]string{
			{"apple", "apple", "fruit"},
			{"banana", "banana", "fruit"},
			{"carrot", "carrot", "vegetable"},
			{"broccoli", "broccoli", "vegetable"},
		})

		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(60, 10))

		// Test initial state shows all items
		tm.WaitContainsStrings(t, []string{"apple", "banana", "carrot", "broccoli"})
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)

		// Note: Filtering is handled internally through the input field
		// We can't directly test SetFilter as it's not exposed
	})

	t.Run("suggestions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column[*table.StaticRow[[]string]]{
			{ID: "name", Title: "Name"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(_ string) tea.Cmd {
				return nil
			},
		})

		m.SetSuggestions([]string{"apple", "banana", "carrot"})
		m.SetItems([][]string{
			{"apple", "apple"},
			{"banana", "banana"},
			{"carrot", "carrot"},
		})

		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(40, 8))
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("selection", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column[*table.StaticRow[[]string]]{
			{ID: "name", Title: "Name"},
			{ID: "value", Title: "Value"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(_ string) tea.Cmd {
				// Selection callback - can be tested by verifying the component works
				return nil
			},
		})

		// Set items - first element is ID, subsequent elements are column data
		m.SetItems([][]string{
			{"item1", "item1", "value1"},
			{"item2", "item2", "value2"},
			{"item3", "item3", "value3"},
		})

		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(50, 8))

		// Test initial state
		tm.WaitContainsStrings(t, []string{"item1", "item2", "item3"})
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)

		// Test that the component renders correctly with data
		// The table component handles selection internally
	})

	t.Run("dimensions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column[*table.StaticRow[[]string]]{
			{ID: "name", Title: "Name"},
			{ID: "description", Title: "Description"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(_ string) tea.Cmd {
				return nil
			},
		})

		// Set items - first element is ID, subsequent elements are column data
		m.SetItems([][]string{
			{"item1", "item1", "description 1"},
			{"item2", "item2", "description 2"},
		})

		// Test with different dimensions
		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(80, 20))
		tm.RequireSnapshotNoANSI(t)

		// Test with small dimensions
		m2 := New(app, Config{
			Columns: columns,
			SelectFunc: func(_ string) tea.Cmd {
				return nil
			},
		})
		m2.SetItems([][]string{
			{"item1", "item1", "description 1"},
			{"item2", "item2", "description 2"},
		})
		tm2 := steep.NewViewModel(t, m2, steep.WithInitialTermSize(40, 10))
		tm2.RequireSnapshotNoANSI(t)
	})
}
