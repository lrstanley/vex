// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package genericcode

import (
	"strings"
	"testing"

	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/steep"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("basic-text-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Test Title", "test content\nline 2\nline 3", "text")
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"test content", "line 2", "line 3"})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("json-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "JSON Test", "{\"name\": \"test\", \"value\": 123}", "json")
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"name", "test", "value", "123"})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("go-code-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Go Code Test", "func test() {\n    return true\n}", "go")
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"func", "test", "return", "true"})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("empty-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Empty Test", "", "text")
		tm := steep.NewComponentHarness(t, m)
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)
	})

	t.Run("zero-dimensions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Zero Test", "test content", "text")
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(0, 0))
		tm.WaitSettleMessages(t).
			RequireDimensions(t, 0, 0)
	})

	t.Run("larger-than-term-size", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app, "Large Test", strings.Repeat("test content\n", 200), "text")
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(100, 15))
		tm.WaitContainsString(t, "test content")
		tm.RequireSnapshotNoANSI(t)
	})
}
