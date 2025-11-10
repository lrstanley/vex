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
	"github.com/lrstanley/vex/internal/cache"
	"github.com/lrstanley/vex/internal/formatter"
)

// TODO: all of this logic is kinda a mess, but it solves the issue for now.

type BorderPosition int

const (
	TopLeftBorder BorderPosition = iota
	TopMiddleBorder
	TopRightBorder
	BottomLeftBorder
	BottomMiddleBorder
	BottomRightBorder
)

type TopLeftBorderEmbed interface {
	TopLeftBorder() string
}

type TopMiddleBorderEmbed interface {
	TopMiddleBorder() string
}

type TopRightBorderEmbed interface {
	TopRightBorder() string
}

type BottomLeftBorderEmbed interface {
	BottomLeftBorder() string
}

type BottomMiddleBorderEmbed interface {
	BottomMiddleBorder() string
}

type BottomRightBorderEmbed interface {
	BottomRightBorder() string
}

type BorderGradient struct {
	TopGradient       []color.Color
	RightGradient     []color.Color
	BottomGradient    []color.Color
	LeftGradient      []color.Color
	TopLeftCorner     color.Color
	TopRightCorner    color.Color
	BottomRightCorner color.Color
	BottomLeftCorner  color.Color
}

var borderGradientCache = cache.New[string, *BorderGradient](10)

// GetBorderGradient returns a border gradient for the given height and width.
func GetBorderGradient(height, width int) *BorderGradient {
	key := fmt.Sprintf("%d-%d", height, width)
	if bg, ok := borderGradientCache.Get(key); ok {
		return bg
	}

	gradient := lipgloss.Blend1D(
		height+width-2, // half of total number of border chars, -2=corners.
		Theme.DialogBorderGradientFromFg(),
		Theme.DialogBorderGradientToFg(),
	)

	// Duplicate the gradient (which is half size), and add it to the end of the gradient in reverse.
	// This allows us to seamlessly blend. E.g. if you just did A->B->C, C wouldn't blend with A when
	// at the end of the gradient. so do A->B->C->C->B->A. It's not perfect, because there are larger
	// consecutive sections of A and C, but looks "good enough".
	rev := slices.Clone(gradient)
	slices.Reverse(rev)
	gradient = append(gradient, rev...)

	bg := &BorderGradient{}

	offset := 0
	getFromOffset := func(size int) []color.Color {
		slice := gradient[offset : offset+size]
		offset += size
		return slice
	}

	bg.TopGradient = getFromOffset(width - 2)
	bg.TopRightCorner = getFromOffset(1)[0]
	bg.RightGradient = getFromOffset(height - 2)
	bg.BottomRightCorner = getFromOffset(1)[0]
	bg.BottomGradient = getFromOffset(width - 2)
	bg.BottomLeftCorner = getFromOffset(1)[0]
	bg.LeftGradient = getFromOffset(height - 2)
	bg.TopLeftCorner = getFromOffset(1)[0]

	// bottom and left gradients are reversed because they are drawn in reverse order.
	slices.Reverse(bg.BottomGradient)
	slices.Reverse(bg.LeftGradient)

	borderGradientCache.Set(key, bg)
	return bg
}

func BorderFromElement(element any) (text map[BorderPosition]string) {
	if element == nil {
		text = make(map[BorderPosition]string)
		return text
	}

	if v, ok := element.(map[BorderPosition]string); ok {
		return v
	}

	text = make(map[BorderPosition]string)

	defaultEmbedStyle := lipgloss.NewStyle().Foreground(Theme.AppFg())

	if v, ok := element.(TopLeftBorderEmbed); ok {
		if vv := v.TopLeftBorder(); vv != "" {
			text[TopLeftBorder] = defaultEmbedStyle.Render(vv)
		}
	}
	if v, ok := element.(TopMiddleBorderEmbed); ok {
		if vv := v.TopMiddleBorder(); vv != "" {
			text[TopMiddleBorder] = defaultEmbedStyle.Render(vv)
		}
	}
	if v, ok := element.(TopRightBorderEmbed); ok {
		if vv := v.TopRightBorder(); vv != "" {
			text[TopRightBorder] = defaultEmbedStyle.Render(vv)
		}
	}
	if v, ok := element.(BottomLeftBorderEmbed); ok {
		if vv := v.BottomLeftBorder(); vv != "" {
			text[BottomLeftBorder] = defaultEmbedStyle.Render(vv)
		}
	}
	if v, ok := element.(BottomMiddleBorderEmbed); ok {
		if vv := v.BottomMiddleBorder(); vv != "" {
			text[BottomMiddleBorder] = defaultEmbedStyle.Render(vv)
		}
	}
	if v, ok := element.(BottomRightBorderEmbed); ok {
		if vv := v.BottomRightBorder(); vv != "" {
			text[BottomRightBorder] = defaultEmbedStyle.Render(vv)
		}
	}

	return text
}

func Border(content string, fg color.Color, element any) string { // nolint:funlen
	width, height := lipgloss.Size(content)

	if height < 2 || width < 2 {
		return ""
	}

	embeddedText := BorderFromElement(element)
	border := lipgloss.RoundedBorder()
	baseStyle := lipgloss.NewStyle().Foreground(Theme.TitleFg())
	bg := GetBorderGradient(height+2, width+2) // +2=borders.

	wrapBrackets := func(text string) string {
		if text != "" {
			return fmt.Sprintf("%s%s%s",
				baseStyle.Render("["),
				formatter.Trunc(text, width-4),
				baseStyle.Render("]"),
			)
		}
		return text
	}

	buildHorizontalBorder := func(
		leftText, middleText, rightText,
		leftCorner, between, rightCorner string,
		gradient []color.Color,
		leftCornerGradient, rightCornerGradient color.Color,
	) string {
		leftText = wrapBrackets(leftText)
		middleText = wrapBrackets(middleText)
		rightText = wrapBrackets(rightText)

		// Calculate length of border between embedded texts.
		remaining := max(0, width-ansi.StringWidth(leftText)-ansi.StringWidth(middleText)-ansi.StringWidth(rightText)-2) // -2=border.
		leftBorderLen := max(0, (width/2)-ansi.StringWidth(leftText)-(ansi.StringWidth(middleText)/2)-1)
		rightBorderLen := max(0, remaining-leftBorderLen)

		// Build gradient border segments.
		var leftBorderSegment, rightBorderSegment strings.Builder

		gradientOffset := 1 + ansi.StringWidth(leftText)
		for range leftBorderLen {
			var c color.Color
			if fg == nil {
				c = gradient[gradientOffset]
				gradientOffset++
			} else {
				c = fg
			}
			leftBorderSegment.WriteString(lipgloss.NewStyle().Foreground(c).Render(between))
		}

		gradientOffset += ansi.StringWidth(middleText)
		for range rightBorderLen {
			var c color.Color
			if fg == nil {
				c = gradient[gradientOffset]
				gradientOffset++
			} else {
				c = fg
			}
			rightBorderSegment.WriteString(lipgloss.NewStyle().Foreground(c).Render(between))
		}

		var leftPaddingBorder, rightPaddingBorder strings.Builder
		var leftPaddingColor, rightPaddingColor color.Color

		if fg == nil {
			// Left padding should use the first gradient color (index 0). Right
			// padding should use the gradient color at the position after all
			// other elements.
			leftPaddingColor = gradient[0]
			rightPaddingColor = gradient[len(gradient)-1]
		} else {
			leftPaddingColor = fg
			rightPaddingColor = fg
		}

		leftPaddingBorder.WriteString(lipgloss.NewStyle().Foreground(leftPaddingColor).Render(between))
		rightPaddingBorder.WriteString(lipgloss.NewStyle().Foreground(rightPaddingColor).Render(between))

		var leftCornerStyle, rightCornerStyle lipgloss.Style
		if fg == nil {
			leftCornerStyle = leftCornerStyle.Foreground(leftCornerGradient)
			rightCornerStyle = rightCornerStyle.Foreground(rightCornerGradient)
		} else {
			leftCornerStyle = leftCornerStyle.Foreground(fg)
			rightCornerStyle = rightCornerStyle.Foreground(fg)
		}

		// Construct the complete border line with padding.
		s := lipgloss.NewStyle().
			Inline(true).
			MaxWidth(width).
			Render(
				leftPaddingBorder.String() +
					leftText +
					leftBorderSegment.String() +
					middleText +
					rightBorderSegment.String() +
					rightText +
					rightPaddingBorder.String(),
			)

		// Add the corners with gradient colors.
		return leftCornerStyle.Render(leftCorner) + s + rightCornerStyle.Render(rightCorner)
	}

	buildVerticalBorders := func(content string, leftGradient, rightGradient []color.Color) string {
		lines := strings.Split(content, "\n")
		result := make([]string, height)
		var leftBorderStyle, rightBorderStyle lipgloss.Style

		for i := range height {
			if fg == nil {
				leftBorderStyle = leftBorderStyle.Foreground(leftGradient[i])
				rightBorderStyle = rightBorderStyle.Foreground(rightGradient[i])
			} else {
				leftBorderStyle = leftBorderStyle.Foreground(fg)
				rightBorderStyle = rightBorderStyle.Foreground(fg)
			}
			result[i] = leftBorderStyle.Render(border.Left) + lines[i] + rightBorderStyle.Render(border.Right)
		}

		return strings.Join(result, "\n")
	}

	// Stack top border, content and horizontal borders, and bottom border.
	return strings.Join([]string{
		buildHorizontalBorder(
			embeddedText[TopLeftBorder],
			embeddedText[TopMiddleBorder],
			embeddedText[TopRightBorder],
			border.TopLeft,
			border.Top,
			border.TopRight,
			bg.TopGradient,
			bg.TopLeftCorner,
			bg.TopRightCorner,
		),
		buildVerticalBorders(content, bg.LeftGradient, bg.RightGradient),
		buildHorizontalBorder(
			embeddedText[BottomLeftBorder],
			embeddedText[BottomMiddleBorder],
			embeddedText[BottomRightBorder],
			border.BottomLeft,
			border.Bottom,
			border.BottomRight,
			bg.BottomGradient,
			bg.BottomLeftCorner,
			bg.BottomRightCorner,
		),
	}, "\n")
}
