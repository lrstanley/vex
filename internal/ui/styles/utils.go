// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"image/color"
	"iter"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/colors"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

// Title returns a styled title string, with a gradient of the from and to
// colors, and the main color for the input string.
//
// Inspired by charmbraclet/crush.
func Title(input string, width int, char string, main, from, to color.Color) string {
	length := lipgloss.Width(input) + 1 // +1 for the space.
	remaining := width - length

	s := lipgloss.NewStyle().Foreground(main)
	if remaining > 0 {
		var out strings.Builder
		clusters := slices.Collect(Clusters(strings.Repeat(char, remaining)))

		for i, c := range colors.BlendLinear1D(len(clusters), from, to) {
			// Allow multiple characters to be rendered.
			if remaining-i == 0 {
				break
			}
			out.WriteString(s.Foreground(c).Render(clusters[i]))
		}

		input = s.Render(input) + " " + out.String()
	}

	return input
}

// Clusters returns an iterator of grapheme clusters from the input string.
func Clusters(input string) iter.Seq[string] {
	return func(yield func(string) bool) {
		gr := uniseg.NewGraphemes(input)
		for gr.Next() {
			if !yield(string(gr.Runes())) {
				return
			}
		}
	}
}

// H returns the cell height of characters in a set of strings. ANSI sequences
// are ignored and characters taller than one cell (such as Chinese characters
// and emojis) are appropriately measured.
//
// You should use this instead of len(string) len([]rune(string) as neither
// will give you accurate results.
func H(input ...string) (h int) {
	for _, s := range input {
		h += lipgloss.Height(s)
	}
	return
}

// W returns the cell width of characters in a set of strings. ANSI sequences
// are ignored and characters wider than one cell (such as Chinese characters
// and emojis) are appropriately measured.
//
// You should use this instead of len(string) len([]rune(string) as neither
// will give you accurate results.
func W(input ...string) (w int) {
	for _, s := range input {
		w += lipgloss.Width(s)
	}
	return
}

// Expand returns a string of v newlines.
func Expand(v int) string {
	if v < 1 {
		return ""
	}

	return strings.Repeat("\n", v)
}

// Trunc truncates a string to a given length, adding a tail to the end if
// the string is longer than the given length. This function is aware of ANSI
// escape codes and will not break them, and accounts for wide-characters (such
// as East-Asian characters and emojis).
// This treats the text as a sequence of graphemes.
func Trunc(s string, length int) string {
	return ansi.Truncate(s, length, IconEllipsis)
}
