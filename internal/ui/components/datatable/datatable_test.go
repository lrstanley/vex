// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package datatable

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/api"
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
	t.Run("basic-table", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, Config[string]{
			NoResultsMsg: "no items found",
			RowFn: func(item string) []string {
				return []string{item, "description"}
			},
		})
		m.SetDimensions(testui.DefaultTermWidth, testui.DefaultTermHeight)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetData([]string{"Name", "Description"}, []string{"item1", "item2"})
		tm.ExpectViewContains(t, "item1", "item2", "Name", "Description")
		tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
		tm.ExpectViewSnapshot(t)
	})

	t.Run("loading-state", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, Config[string]{
			RowFn: func(item string) []string {
				return []string{item}
			},
		})
		m.SetDimensions(testui.DefaultTermWidth, testui.DefaultTermHeight)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "loading")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("no-results", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, Config[string]{
			NoResultsMsg: "no items found",
			RowFn: func(item string) []string {
				return []string{item}
			},
		})
		m.SetDimensions(testui.DefaultTermWidth, testui.DefaultTermHeight)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetData([]string{"Name"}, []string{})
		tm.ExpectViewContains(t, "no items found")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("filtered-no-results", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, Config[string]{
			NoResultsFilterMsg: "no results for %q",
			RowFn: func(item string) []string {
				return []string{item}
			},
		})
		m.SetDimensions(testui.DefaultTermWidth, testui.DefaultTermHeight)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetData([]string{"Name"}, []string{"item1", "item2"})
		m.SetFilter("nonexistent")
		tm.ExpectViewContains(t, "no results for \"nonexistent\"")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, Config[string]{
			RowFn: func(item string) []string {
				return []string{item}
			},
		})
		m.SetDimensions(0, 0)
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(0, 0))
		m.SetData([]string{"Name"}, []string{"item1"})
		tm.ExpectViewSnapshot(t)
	})
}
