// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package loader

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/ui/testui"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
	os.Exit(v)
}

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-loader", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetHeight(testui.DefaultTermHeight)
		m.SetWidth(testui.DefaultTermWidth)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "loading")
		tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetHeight(0)
		m.SetWidth(0)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})

	t.Run("small-dimensions", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetHeight(5)
		m.SetWidth(20)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewContains(t, "loading")
		tm.ExpectViewSnapshot(t)
	})
}
