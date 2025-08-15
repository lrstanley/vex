// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const TruncateEllipsis = "…" // Should be 1 character wide.

// Trunc truncates a string to a given length, adding a tail to the end if the
// string is longer than the given length. This function is aware of ANSI escape
// codes and will not break them, and accounts for wide-characters (such as
// East-Asian characters and emojis). This treats the text as a sequence of
// graphemes.
func Trunc(s string, length int) string {
	return ansi.Truncate(s, length, TruncateEllipsis)
}

func TruncLeft(s string, length int) string {
	return ansi.TruncateLeft(s, length, TruncateEllipsis)
}

// TruncMultiline is similar to [Trunc], but it truncates each line of a
// multiline string separately.
func TruncMultiline(s string, length int) string {
	out := strings.Split(s, "\n")
	for i := range out {
		out[i] = ansi.Truncate(out[i], length, TruncateEllipsis)
	}
	return strings.Join(out, "\n")
}

func TruncLeftMultiline(s string, length int) string {
	out := strings.Split(s, "\n")
	for i := range out {
		out[i] = ansi.TruncateLeft(out[i], length, TruncateEllipsis)
	}
	return strings.Join(out, "\n")
}

func CutMultiline(s string, left, right int) string {
	out := strings.Split(s, "\n")
	for i := range out {
		out[i] = ansi.Cut(out[i], left, right)
	}
	return strings.Join(out, "\n")
}

const ANSIReset = "\x1b[m"

// TruncReset removes the ANSI reset sequence from a string.
func TruncReset(s string) string {
	return strings.ReplaceAll(s, ANSIReset, "")
}
