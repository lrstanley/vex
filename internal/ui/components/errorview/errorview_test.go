// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package errorview

import (
	"errors"
	"strings"
	"testing"

	"github.com/lrstanley/x/charm/steep"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("2 errors", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetErrors(
			errors.New("test error 1"),
			errors.New("test error 2"),
		)
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"2 errors", "test error 1", "test error 2"})
		tm.RequireDimensions(t, m.GetWidth(), m.GetHeight()).
			RequireSnapshotNoANSI(t)
	})

	t.Run("too-many-errors", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetErrors(errors.New(strings.Repeat("test error\n", 100)))
		tm := steep.NewComponentHarness(t, m)
		tm.WaitContainsStrings(t, []string{"1 errors", "error(s) not shown"})
		tm.RequireDimensions(t, m.GetWidth(), m.GetHeight()).
			RequireSnapshotNoANSI(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetErrors(errors.New("test error"))
		tm := steep.NewComponentHarness(t, m, steep.WithInitialTermSize(0, 0))
		tm.WaitSettleMessages(t).
			RequireDimensions(t, 0, 0)
	})
}
