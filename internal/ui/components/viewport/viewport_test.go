// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package viewport

import (
	"strings"
	"testing"

	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/steep"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-viewport", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetContent("test content\nline 2\nline 3")
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"test content", "line 2", "line 3"})
		tm.RequireDimensions(t, m.GetWidth(), m.GetHeight()).
			RequireSnapshotNoANSI(t)
	})

	t.Run("empty-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetContent("")
		tm := steep.NewComponentHarness(t, m)
		tm.WaitSettleMessages(t).RequireSnapshotNoANSI(t)
	})

	t.Run("json-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		data := map[string]any{
			"name":  "test",
			"value": 123,
		}
		err := m.SetJSON(data)
		if err != nil {
			t.Fatalf("failed to set JSON: %v", err)
		}
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"name", "test", "value", "123"})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("code-content", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetCode("func test() {\n    return true\n}", "go")
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"func", "test", "return", "true"})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetContent("test content")
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(0, 0))
		tm.WaitSettleMessages(t).
			RequireDimensions(t, 0, 0)
	})

	t.Run("content-larger-than-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetContent(strings.Repeat("test content\n", steep.DefaultTermHeight*2))
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsString(t, "test content")
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("content-larger-than-height-at-bottom", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.SetContent(strings.TrimSpace(strings.Repeat("test content\n", steep.DefaultTermHeight*2)))
		tm := steep.NewComponentHarness(t, m)
		steep.Mutate(t, tm, func(m *Model) *Model {
			m.GotoBottom()
			return m
		})
		tm.WaitContainsString(t, "test content")
		tm.RequireSnapshotNoANSI(t)
	})
}
