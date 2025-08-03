// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package viewport

import (
	"os"
	"strings"
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
	t.Run("basic-viewport", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetHeight(testui.DefaultTermHeight)
		m.SetWidth(testui.DefaultTermWidth)
		m.SetContent("test content\nline 2\nline 3")
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "test content", "line 2", "line 3")
		tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
		tm.ExpectViewSnapshot(t)
	})

	t.Run("empty-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetHeight(testui.DefaultTermHeight)
		m.SetWidth(testui.DefaultTermWidth)
		m.SetContent("")
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})

	t.Run("json-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetHeight(testui.DefaultTermHeight)
		m.SetWidth(testui.DefaultTermWidth)
		data := map[string]any{
			"name":  "test",
			"value": 123,
		}
		err := m.SetJSON(data)
		if err != nil {
			t.Fatalf("failed to set JSON: %v", err)
		}
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "name", "test", "value", "123")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("code-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetHeight(testui.DefaultTermHeight)
		m.SetWidth(testui.DefaultTermWidth)
		m.SetCode("func test() {\n    return true\n}", "go")
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "func", "test", "return", "true")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetContent("test content")
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(0, 0))
		tm.ExpectViewSnapshot(t)
	})

	t.Run("content-larger-than-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetContent(strings.Repeat("test content\n", testui.DefaultTermHeight*2))
		tm.ExpectContains(t, "test content")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("content-larger-than-height-at-bottom", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetContent(strings.TrimSpace(strings.Repeat("test content\n", testui.DefaultTermHeight*2)))
		m.GotoBottom()
		tm.ExpectContains(t, "test content")
		tm.ExpectViewSnapshot(t)
	})
}
