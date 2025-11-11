// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/x/charm/formatter"
)

// Title returns a styled title string, with a gradient of the from and to
// colors, and the main color for the input string.
//
// Inspired by charmbraclet/crush.
func Title(input string, width int, char string, main, from, to color.Color) string {
	length := ansi.StringWidth(input)
	remaining := width - length

	if remaining < 1 {
		return formatter.Trunc(input, width)
	}

	remaining-- // -1 for the space.

	s := lipgloss.NewStyle().Foreground(main)
	if remaining > 0 {
		var out strings.Builder
		clusters := slices.Collect(formatter.Clusters(strings.Repeat(char, remaining)))

		for i, c := range lipgloss.Blend1D(len(clusters), from, to) {
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
	return h
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
	return w
}

// Expand returns a string of v newlines.
func Expand(v int) string {
	if v < 1 {
		return ""
	}

	return strings.Repeat("\n", v)
}

// Pluralize returns a string with the count and the singular or plural form of
// the word, primarily used for things like border based titles.
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}
