// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
)

const ScrollbarWidth = 1

func (tc *ThemeConfig) Scrollbar(height, total, visible, offset int) string {
	if total == visible {
		return strings.TrimRight(strings.Repeat(" \n", height), "\n")
	}

	ratio := float64(height) / float64(total)
	thumbHeight := max(1, int(math.Round(float64(visible)*ratio)))
	thumbOffset := max(0, min(height-thumbHeight, int(math.Round(float64(offset)*ratio))))

	track := lipgloss.NewStyle().Foreground(tc.ScrollbarTrackFg()).Render(IconScrollbar)
	thumb := lipgloss.NewStyle().Foreground(tc.ScrollbarThumbFg()).Render(IconScrollbar)

	return strings.TrimRight(
		strings.Repeat(track+"\n", thumbOffset)+
			strings.Repeat(thumb+"\n", thumbHeight)+
			strings.Repeat(track+"\n", max(0, height-thumbOffset-thumbHeight)),
		"\n",
	)
}
