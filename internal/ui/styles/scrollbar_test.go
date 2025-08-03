// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/ui/testui"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true}) //nolint:errcheck
	os.Exit(v)
}

func TestScrollbar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		height  int
		total   int
		visible int
		offset  int
	}{
		{
			name:    "no-scroll-needed",
			height:  10,
			total:   10,
			visible: 10,
			offset:  0,
		},
		{
			name:    "large-content-at-top",
			height:  10,
			total:   100,
			visible: 10,
			offset:  0,
		},
		{
			name:    "large-content-at-middle",
			height:  10,
			total:   100,
			visible: 10,
			offset:  50,
		},
		{
			name:    "large-content-at-bottom",
			height:  10,
			total:   100,
			visible: 10,
			offset:  90,
		},
		{
			name:    "medium-content",
			height:  10,
			total:   50,
			visible: 10,
			offset:  0,
		},
		{
			name:    "zero-height",
			height:  0,
			total:   100,
			visible: 10,
			offset:  0,
		},
		{
			name:    "zero-total",
			height:  10,
			total:   0,
			visible: 10,
			offset:  0,
		},
		{
			name:    "offset-exceeds-bounds",
			height:  10,
			total:   100,
			visible: 10,
			offset:  200,
		},
		{
			name:    "visible-larger-than-total",
			height:  10,
			total:   5,
			visible: 10,
			offset:  0,
		},
		{
			name:    "snapshot-test",
			height:  20,
			total:   100,
			visible: 20,
			offset:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testui.ExpectSnapshotNonANSI(t, Scrollbar(tt.height, tt.total, tt.visible, tt.offset, IconScrollbar, " "))
		})
	}
}
