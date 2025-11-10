// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package dialogselector

import (
	"os"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/vex/internal/ui/testui"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true}) //nolint:errcheck
	os.Exit(v)
}

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-selector", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		m := New(app, Config{
			Columns: []*table.Column{
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(testui.DefaultTermWidth, testui.DefaultTermHeight))
		tm.ExpectViewContains(t, "description 1")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("empty-list", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		m := New(app, Config{
			Columns: []*table.Column{
				{ID: "name", Title: "Name"},
				{ID: "description", Title: "Description"},
			},
			SelectFunc: func(_ string) tea.Cmd { return nil },
		})

		m.SetItems([][]string{})

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(testui.DefaultTermWidth, testui.DefaultTermHeight))
		tm.ExpectViewContains(t, "no results found")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		m := New(app, Config{
			Columns: []*table.Column{
				{ID: "name", Title: "Name"},
				{ID: "description", Title: "Description"},
			},
			SelectFunc: func(id string) tea.Cmd { return nil },
		})

		// Set items after creation
		// First element is the ID, subsequent elements are the column data
		m.SetItems([][]string{{"item1", "item1", "description"}})

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(0, 0))
		tm.ExpectViewSnapshot(t)
	})
}

func TestDialogSelectorFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("filtering", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column{
			{ID: "name", Title: "Name"},
			{ID: "category", Title: "Category"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(id string) tea.Cmd {
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(60, 10))

		// Test initial state shows all items
		tm.ExpectViewContains(t, "apple", "banana", "carrot", "broccoli")
		tm.ExpectViewSnapshot(t)

		// Note: Filtering is handled internally through the input field
		// We can't directly test SetFilter as it's not exposed
	})

	t.Run("suggestions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column{
			{ID: "name", Title: "Name"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(id string) tea.Cmd {
				return nil
			},
		})

		m.SetSuggestions([]string{"apple", "banana", "carrot"})
		m.SetItems([][]string{
			{"apple", "apple"},
			{"banana", "banana"},
			{"carrot", "carrot"},
		})

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(40, 8))
		tm.ExpectViewSnapshot(t)
	})

	t.Run("selection", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column{
			{ID: "name", Title: "Name"},
			{ID: "value", Title: "Value"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(id string) tea.Cmd {
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

		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(50, 8))

		// Test initial state
		tm.ExpectViewContains(t, "item1", "item2", "item3")
		tm.ExpectViewSnapshot(t)

		// Test that the component renders correctly with data
		// The table component handles selection internally
	})

	t.Run("dimensions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)

		columns := []*table.Column{
			{ID: "name", Title: "Name"},
			{ID: "description", Title: "Description"},
		}

		m := New(app, Config{
			Columns: columns,
			SelectFunc: func(id string) tea.Cmd {
				return nil
			},
		})

		// Set items - first element is ID, subsequent elements are column data
		m.SetItems([][]string{
			{"item1", "item1", "description 1"},
			{"item2", "item2", "description 2"},
		})

		// Test with different dimensions
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(80, 20))
		tm.ExpectViewSnapshot(t)

		// Test with small dimensions
		tm2 := testui.NewNonRootModel(t, m, false, testui.WithTermSize(40, 10))
		tm2.ExpectViewSnapshot(t)
	})
}
