// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package testui

import (
	"bytes"
	"os"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/gkampitakis/go-snaps/snaps"
)

var SnapConfig = snaps.WithConfig(
	snaps.Dir("testdata"),
)

// ExpectSnapshotProfile is a helper function that will create snapshots for
// visually identifying the output of a view, with a specific color profile,
// which can be used to automatically downgrade color data.
func ExpectSnapshotProfile[T []byte | string](tb testing.TB, out T, profile colorprofile.Profile) {
	tb.Helper()

	buf := &bytes.Buffer{}

	w := &colorprofile.Writer{
		Forward: buf,
		Profile: profile,
	}

	_, err := w.Write([]byte(out))
	if err != nil {
		tb.Fatalf("failed to write view: %v", err)
	}

	ExpectSnapshot(tb, buf.String())
}

// ExpectSnapshot is a helper function that will create snapshots for visually
// identifying the output of a view.
func ExpectSnapshot[T []byte | string](tb testing.TB, out T) {
	tb.Helper()
	SnapConfig.MatchSnapshot(tb, out)
}

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
	os.Exit(v)
}
