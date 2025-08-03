// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"image/color"
	"testing"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/ui/testui"
)

func TestBorder(t *testing.T) {
	t.Parallel()

	// Define test colors
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	x, y := 60, 10

	s := lipgloss.NewStyle().
		Width(x).
		Height(y).
		MaxWidth(x).
		MaxHeight(y)

	tests := []struct {
		name    string
		content string
		fg      color.Color
		element any
	}{
		{
			name:    "basic-border-with-static-color",
			content: s.Render("Hello\nWorld"),
			fg:      red,
		},
		{
			name:    "basic-border-with-gradient",
			content: s.Render("Hello\nWorld"), // Use gradient
		},
		{
			name:    "small-content",
			content: s.Render("Hi"),
			fg:      blue,
		},
		{
			name:    "large-content",
			content: s.Render("This is a longer content\nwith multiple lines\nfor testing borders"),
			fg:      green,
		},
		{
			name:    "single-line-content",
			content: s.Render("Single line content"),
			fg:      red,
		},
		{
			name:    "empty-content",
			content: s.Render(""),
			fg:      blue,
		},
		{
			name:    "very-small-content",
			content: s.Render("A"),
			fg:      green,
		},
		{
			name:    "top-left-embedded-text",
			content: s.Render("Content\nwith embedded text"),
			fg:      red,
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopLeftBorder: "TL",
			},
		},
		{
			name:    "top-middle-embedded-text",
			content: s.Render("Content\nwith embedded text"),
			fg:      blue,
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopMiddleBorder: "TM",
			},
		},
		{
			name:    "top-right-embedded-text",
			content: s.Render("Content\nwith embedded text"),
			fg:      green,
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopRightBorder: "TR",
			},
		},
		{
			name:    "bottom-left-embedded-text",
			content: s.Render("Content\nwith embedded text"),
			fg:      red,
			element: map[BorderPosition]string{ //nolint:exhaustive
				BottomLeftBorder: "BL",
			},
		},
		{
			name:    "bottom-middle-embedded-text",
			content: s.Render("Content\nwith embedded text"),
			fg:      blue,
			element: map[BorderPosition]string{ //nolint:exhaustive
				BottomMiddleBorder: "BM",
			},
		},
		{
			name:    "bottom-right-embedded-text",
			content: s.Render("Content\nwith embedded text"),
			fg:      green,
			element: map[BorderPosition]string{ //nolint:exhaustive
				BottomRightBorder: "BR",
			},
		},
		{
			name:    "all-embedded-texts",
			content: s.Render("Content\nwith all embedded texts"),
			fg:      red,
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopLeftBorder:      "TL",
				TopMiddleBorder:    "TM",
				TopRightBorder:     "TR",
				BottomLeftBorder:   "BL",
				BottomMiddleBorder: "BM",
				BottomRightBorder:  "BR",
			},
		},
		{
			name:    "all-embedded-texts-with-gradient",
			content: s.Render("Content\nwith all embedded texts"), // Use gradient
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopLeftBorder:      "TL",
				TopMiddleBorder:    "TM",
				TopRightBorder:     "TR",
				BottomLeftBorder:   "BL",
				BottomMiddleBorder: "BM",
				BottomRightBorder:  "BR",
			},
		},
		{
			name:    "manual-embedded-text",
			content: s.Render("Content\nwith manual embedded text"),
			fg:      blue,
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopLeftBorder:      "MANUAL_TL",
				TopMiddleBorder:    "MANUAL_TM",
				TopRightBorder:     "MANUAL_TR",
				BottomLeftBorder:   "MANUAL_BL",
				BottomMiddleBorder: "MANUAL_BM",
				BottomRightBorder:  "MANUAL_BR",
			},
		},
		{
			name:    "manual-embedded-text-with-gradient",
			content: s.Render("Content\nwith manual embedded text"), // Use gradient
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopLeftBorder:      "MANUAL_TL",
				TopMiddleBorder:    "MANUAL_TM",
				TopRightBorder:     "MANUAL_TR",
				BottomLeftBorder:   "MANUAL_BL",
				BottomMiddleBorder: "MANUAL_BM",
				BottomRightBorder:  "MANUAL_BR",
			},
		},
		{
			name:    "wide-content",
			content: s.Render("This is a very wide content that should test border rendering with long lines"),
			fg:      red,
		},
		{
			name:    "tall-content",
			content: s.Render("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10"),
			fg:      blue,
		},
		{
			name:    "wide-and-tall-content",
			content: s.Render("This is a very wide line that should test border rendering\nLine 2\nLine 3\nLine 4\nLine 5"),
			fg:      green,
		},
		{
			name:    "content-with-special-chars",
			content: s.Render("Content with special chars: !@#$%^&*()\nAnd more: []{}|\\:;\"'<>?,./"),
			fg:      red,
		},
		{
			name:    "unicode-content",
			content: s.Render("Unicode content: 🚀🌟🎉\nMore unicode: 🎨🎭🎪"),
			fg:      blue,
		},
		{
			name:    "unicode-content-with-gradient",
			content: s.Render("Unicode content: 🚀🌟🎉\nMore unicode: 🎨🎭🎪"), // Use gradient
		},
		{
			name:    "mixed-unicode-and-text",
			content: s.Render("Mixed: Hello 🚀 World 🌟\nMore: Test 🎉 Content"),
			fg:      green,
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopMiddleBorder: "TM",
			},
		},
		{
			name:    "empty-lines-content",
			content: s.Render("First line\n\nThird line\n\nFifth line"),
			fg:      red,
		},
		{
			name:    "single-char-lines",
			content: s.Render("A\nB\nC\nD\nE"),
			fg:      blue,
		},
		{
			name:    "very-wide-single-line",
			content: s.Render("This is an extremely wide line that should test how the border handles very long content without line breaks"),
			fg:      green,
		},
		{
			name:    "very-wide-single-line-with-gradient",
			content: s.Render("This is an extremely wide line that should test how the border handles very long content without line breaks"), // Use gradient
		},
		{
			name:    "partial-embedded-texts",
			content: s.Render("Content\nwith partial embedded texts"),
			fg:      red,
			element: map[BorderPosition]string{ //nolint:exhaustive
				TopLeftBorder:    "TL",
				TopRightBorder:   "TR",
				BottomLeftBorder: "BL",
			},
		},
		{
			name:    "empty-embedded-texts",
			content: s.Render("Content\nwith empty embedded texts"),
			fg:      blue,
			element: map[BorderPosition]string{
				TopLeftBorder:      "",
				TopMiddleBorder:    "",
				TopRightBorder:     "",
				BottomLeftBorder:   "",
				BottomMiddleBorder: "",
				BottomRightBorder:  "",
			},
		},
		{
			name:    "long-embedded-texts",
			content: s.Render("Content\nwith long embedded texts"),
			fg:      green,
			element: map[BorderPosition]string{
				TopLeftBorder:      "VERY_LONG_TOP_LEFT_TEXT",
				TopMiddleBorder:    "VERY_LONG_TOP_MIDDLE_TEXT",
				TopRightBorder:     "VERY_LONG_TOP_RIGHT_TEXT",
				BottomLeftBorder:   "VERY_LONG_BOTTOM_LEFT_TEXT",
				BottomMiddleBorder: "VERY_LONG_BOTTOM_MIDDLE_TEXT",
				BottomRightBorder:  "VERY_LONG_BOTTOM_RIGHT_TEXT",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Border(tt.content, tt.fg, tt.element)
			testui.ExpectSnapshotNonANSI(t, result)
		})
	}
}

func TestBorderEdgeCases(t *testing.T) {
	t.Parallel()

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	tests := []struct {
		name        string
		content     string
		fg          color.Color
		element     any
		expectEmpty bool
	}{
		{
			name:        "too-small-height",
			content:     "A",
			fg:          red,
			expectEmpty: true,
		},
		{
			name:        "too-small-width",
			content:     "A",
			fg:          red,
			expectEmpty: true,
		},
		{
			name:        "single-char",
			content:     "X",
			fg:          red,
			expectEmpty: true,
		},
		{
			name:        "empty-string",
			content:     "",
			fg:          red,
			expectEmpty: true,
		},
		{
			name:    "nil-fg-color",
			content: "Hello\nWorld",
		},
		{
			name:    "nil-element",
			content: "Hello\nWorld",
			fg:      red,
		},
		{
			name:    "nil-embedded-text",
			content: "Hello\nWorld",
			fg:      red,
		},
		{
			name:    "nil-mock-element",
			content: "Hello\nWorld",
			fg:      red,
		},
		{
			name:    "empty-mock-element",
			content: "Hello\nWorld",
			fg:      red,
			element: map[BorderPosition]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Border(tt.content, tt.fg, tt.element)

			if tt.expectEmpty {
				if result != "" {
					t.Errorf("Expected empty result, got: %q", result)
				}
			} else {
				if result == "" {
					t.Errorf("Expected non-empty result, got empty")
				}
				testui.ExpectSnapshotNonANSI(t, result)
			}
		})
	}
}
