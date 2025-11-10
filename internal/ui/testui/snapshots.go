// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package testui

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	"github.com/charmbracelet/colorprofile"
	"github.com/gkampitakis/go-snaps/snaps"
)

var SnapConfig = snaps.WithConfig(
	snaps.Dir("testdata"),
)

// ExpectSnapshotProfile is a helper function that will create snapshots for
// visually identifying the output of a view, with a specific color profile,
// which can be used to automatically downgrade color data.
func ExpectSnapshotProfile(tb testing.TB, out string, profile colorprofile.Profile) {
	tb.Helper()

	buf := &bytes.Buffer{}

	w := &colorprofile.Writer{
		Forward: buf,
		Profile: profile,
	}

	_, err := w.WriteString(out)
	if err != nil {
		tb.Fatalf("failed to write view: %v", err)
	}

	ExpectSnapshot(tb, buf.String())
}

// ExpectSnapshot is a helper function that will create snapshots for visually
// identifying the output of a view.
func ExpectSnapshot(tb testing.TB, out string) {
	tb.Helper()

	out = stripSpinner(out)

	SnapConfig.MatchSnapshot(tb, out)
}

func ExpectSnapshotNonANSI(tb testing.TB, out string) {
	tb.Helper()

	out = stripSpinner(out)

	SnapConfig.MatchSnapshot(tb, Strip(out))
}

func stripSpinner(out string) string {
	for _, s := range spinner.MiniDot.Frames {
		out = strings.ReplaceAll(out, s, "<spinner>")
	}
	return out
}
