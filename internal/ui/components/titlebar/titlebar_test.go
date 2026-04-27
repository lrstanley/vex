// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package titlebar

import (
	"testing"

	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/steep"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-infobar", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"mock-page", "help"})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(0, 0))
		tm.RequireDimensions(t, 0, 0)
	})

	t.Run("small-dimensions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(40, 3))
		tm.WaitContainsStrings(t, []string{"mock-page", "help"})
		tm.RequireSnapshotNoANSI(t)
	})
}
