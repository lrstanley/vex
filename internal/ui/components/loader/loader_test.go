// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package loader

import (
	"testing"

	"github.com/lrstanley/x/charm/steep"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-loader", func(t *testing.T) {
		t.Parallel()
		m := New()
		tm := steep.NewViewModel(t, m)
		tm.WaitContainsString(t, "loading")
		tm.ExpectDimensions(t, m.GetWidth(), m.GetHeight()).RequireSnapshotNoANSI(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		m := New()
		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(0, 0))
		tm.WaitSettleView(t).ExpectDimensions(t, 0, 0)
	})

	t.Run("small-dimensions", func(t *testing.T) {
		t.Parallel()
		m := New()
		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(20, 5))
		tm.WaitContainsString(t, "loading")
		tm.RequireSnapshotNoANSI(t)
	})
}
