// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package errorview

import (
	"errors"
	"strings"
	"testing"

	"github.com/lrstanley/vex/internal/ui/testui"
)

func TestMain(m *testing.M) {
	testui.TestMain(m)
}

func TestNew(t *testing.T) {
	t.Run("2 errors", func(t *testing.T) {
		m := New()
		m.SetHeight(testui.DefaultTermHeight)
		m.SetWidth(testui.DefaultTermWidth)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetErrors(
			errors.New("test error 1"),
			errors.New("test error 2"),
		)
		tm.ExpectViewContains(t, "2 errors", "test error 1", "test error 2")
		tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
		tm.ExpectViewSnapshot(t)
	})

	t.Run("too-many-errors", func(t *testing.T) {
		m := New()
		m.SetHeight(testui.DefaultTermHeight)
		m.SetWidth(testui.DefaultTermWidth)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetErrors(
			errors.New(strings.Repeat("test error\n", 100)),
		)
		tm.ExpectViewContains(t, "1 errors", "error(s) not shown")
		tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		m := New()
		m.SetHeight(0)
		m.SetWidth(0)
		tm := testui.NewNonRootModel(t, m, false)
		m.SetErrors(errors.New("test error"))
		tm.ExpectViewSnapshot(t)
	})
}
