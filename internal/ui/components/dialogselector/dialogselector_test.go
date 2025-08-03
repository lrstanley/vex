// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package dialogselector

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
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

type mockListable struct {
	suggestions []string
	data        [][]string
}

func (m *mockListable) Suggestions() []string {
	return m.suggestions
}

func (m *mockListable) Len() int {
	return len(m.data)
}

func (m *mockListable) GetData() ([]string, [][]string) {
	columns := []string{"Name", "Description"}
	return columns, m.data
}

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-selector", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		list := &mockListable{
			suggestions: []string{"item1", "item2", "item3"},
			data: [][]string{
				{"item1", "description 1"},
				{"item2", "description 2"},
				{"item3", "description 3"},
			},
		}
		m := New(app, Config{
			List:              list,
			FilterPlaceholder: "type to filter",
			SelectFunc: func(row []string) tea.Cmd {
				return nil
			},
		})
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(testui.DefaultTermWidth, testui.DefaultTermHeight))
		tm.ExpectContains(t, "description 1")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("empty-list", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		list := &mockListable{
			suggestions: []string{},
			data:        [][]string{},
		}
		m := New(app, Config{
			List: list,
			SelectFunc: func(row []string) tea.Cmd {
				return nil
			},
		})
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(testui.DefaultTermWidth, testui.DefaultTermHeight))
		tm.ExpectContains(t, "no results found")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		list := &mockListable{
			suggestions: []string{"item1"},
			data:        [][]string{{"item1", "description"}},
		}
		m := New(app, Config{
			List: list,
			SelectFunc: func(row []string) tea.Cmd {
				return nil
			},
		})
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(0, 0))
		tm.ExpectViewSnapshot(t)
	})
}
