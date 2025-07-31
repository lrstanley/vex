// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package errorview

import (
	"errors"
	"testing"

	"github.com/lrstanley/vex/internal/ui/testui"
)

func TestNew(t *testing.T) {
	m := New()
	m.SetHeight(testui.DefaultTermHeight)
	m.SetWidth(testui.DefaultTermWidth)
	m.SetErrors(
		errors.New("test error 1"),
		errors.New("test error 2"),
	)

	tm := testui.NewNonRootModel(t, m, false)
	tm.ExpectContains(t, "2 errors", "test error 1", "test error 2")
	tm.ExpectViewDimensions(t, m.GetWidth(), m.GetHeight())
	tm.ExpectViewSnapshot(t)
	tm.ExpectViewSnapshot(t)

	// TODO:
	//   - validate when too many errors show.
	//   - test 0 width/height dims.
}
