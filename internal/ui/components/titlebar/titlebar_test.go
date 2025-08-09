// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package titlebar

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
	t.Run("basic-infobar", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.Height = testui.DefaultTermHeight
		m.Width = testui.DefaultTermWidth
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "mock-page", "help")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.Height = 0
		m.Width = 0
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "mock-page", "help")
		tm.ExpectViewSnapshot(t)
	})

	t.Run("small-dimensions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		m.Height = 3
		m.Width = 40
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "mock-page", "help")
		tm.ExpectViewSnapshot(t)
	})
}
