// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package genericcode

import (
	"os"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/testui"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true}) //nolint:errcheck
	os.Exit(v)
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("basic-text-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Test Title", "test content\nline 2\nline 3", "text")
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectContains(t, "test content", "line 2", "line 3")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("json-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "JSON Test", "{\"name\": \"test\", \"value\": 123}", "json")
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectContains(t, "name", "test", "value", "123")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("go-code-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Go Code Test", "func test() {\n    return true\n}", "go")
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectContains(t, "func", "test", "return", "true")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("empty-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Empty Test", "", "text")
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})

	t.Run("zero-dimensions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Zero Test", "test content", "text")
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(0, 0))
		tm.ExpectViewSnapshot(t)
	})

	t.Run("larger-than-term-size", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Large Test", strings.Repeat("test content\n", 200), "text")
		tm := testui.NewNonRootModel(t, m, false, testui.WithTermSize(100, 15))
		tm.ExpectContains(t, "test content")
		tm.ExpectViewSnapshot(t)
	})
}
