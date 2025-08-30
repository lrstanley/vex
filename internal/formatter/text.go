// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"iter"
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
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

// TruncPath dynamically truncates a path to a given length, prioritizing keeping
// both start and end segments when possible.
func TruncPath(s string, length int) string {
	if length <= 0 {
		return ""
	}

	sw := ansi.StringWidth(s)
	if sw <= length {
		return s
	}

	parts := slices.DeleteFunc(strings.SplitAfter(s, "/"), func(s string) bool {
		return s == ""
	})

	if len(parts) == 1 {
		return Trunc(parts[0], length)
	}

	// Split parts into left and right halves, as close as possible to the center.
	var left, right []string

	for i := range parts {
		if i >= len(parts)/2 {
			left = parts[:i]
			right = parts[i:]
			break
		}
	}

	if len(left) == 0 || len(right) == 0 {
		return Trunc(s, length)
	}

	var w, ellipsisWidth int

	for sw+ellipsisWidth > length {
		if len(left) >= len(right) {
			if len(left) == 0 {
				break
			}

			// Delete the last part of the left side.
			w = ansi.StringWidth(left[len(left)-1])
			left = left[:len(left)-1]
			sw -= w
		} else {
			if len(right) <= 1 {
				break
			}

			// Delete the first part of the right side.
			w = ansi.StringWidth(right[0])
			right = right[1:]
			sw -= w
		}

		if ellipsisWidth == 0 {
			ellipsisWidth = 2
		}
	}

	if len(left) == 0 && len(right) == 0 {
		return TruncateEllipsis
	}

	if len(left)+len(right) != len(parts) {
		return Trunc(strings.Join(left, "")+TruncateEllipsis+"/"+strings.Join(right, ""), length)
	}

	return Trunc(
		strings.Join(parts, ""),
		length,
	)
}

// TruncMaybePath truncates a string similar to [Trunc], but if one of the parts
// of the string looks like a path, it will be truncated using [TruncPath].
func TruncMaybePath(s string, length int) string {
	w := ansi.StringWidth(s)
	if w <= length {
		return s
	}

	parts := strings.Split(s, " ")
	pathi := -1

	for i := range parts {
		if strings.Contains(parts[i], "/") && ansi.StringWidth(parts[i]) > 1 {
			pathi = i
			break
		}
	}

	if pathi == -1 {
		return Trunc(s, length)
	}

	before := strings.Join(parts[:pathi], " ")
	beforew := ansi.StringWidth(before)
	after := strings.Join(parts[pathi+1:], " ")
	afterw := ansi.StringWidth(after)

	var out strings.Builder

	if beforew > 0 {
		out.WriteString(before)
		out.WriteString(" ")
	}

	maxPathWidth := length - beforew - afterw

	if beforew > 0 {
		maxPathWidth--
	}

	if maxPathWidth > 0 {
		maxPathWidth--
	}

	out.WriteString(TruncPath(parts[pathi], maxPathWidth))

	if afterw > 0 {
		out.WriteString(" ")
		out.WriteString(after)
	}

	return out.String()
}

// Clusters returns an iterator of grapheme clusters from the input string.
func Clusters(input string) iter.Seq[string] {
	return func(yield func(string) bool) {
		gr := uniseg.NewGraphemes(input)
		for gr.Next() {
			if !yield(gr.Str()) {
				return
			}
		}
	}
}
