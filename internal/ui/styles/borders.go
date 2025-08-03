// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"fmt"
	"image/color"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
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

func rotate[T any, I int | int64](arr []T, k I) {
	n := len(arr)
	k %= I(n)
	if k < 0 {
		k += I(n)
	}
	slices.Reverse(arr[:k])
	slices.Reverse(arr[k:])
	slices.Reverse(arr)
}

var (
	borderRotation    = atomic.Int64{}
	borderRotationFPS = 25
)

type BorderRotationTickMsg struct {
	Current int64
}

func BorderRotationTick(msg tea.Msg) tea.Cmd {
	// This is purely exploratory. Current design is terrible and uses a lot of CPU,
	// so may explore with a more optimized approach in the future.
	//
	//	rotate(gradient, borderRotation.Load())
	//
	v, ok := msg.(BorderRotationTickMsg)
	if msg != nil && !ok {
		return nil
	}

	if msg != nil && v.Current != borderRotation.Load() {
		return nil
	}

	return tea.Tick(time.Second/time.Duration(borderRotationFPS), func(_ time.Time) tea.Msg {
		return BorderRotationTickMsg{Current: borderRotation.Add(-3)}
	})
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

	defaultEmbedStyle := lipgloss.NewStyle().Foreground(Theme.Fg())

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
	height := lipgloss.Height(content)
	width := lipgloss.Width(content)

	if height < 2 || width < 2 {
		return ""
	}

	embeddedText := BorderFromElement(element)

	border := lipgloss.RoundedBorder()
	baseStyle := lipgloss.NewStyle().Foreground(Theme.DialogTitleFg())

	var topGradient, rightGradient, bottomGradient, leftGradient []color.Color
	var topLeftCornerGradient, topRightCornerGradient, bottomLeftCornerGradient, bottomRightCornerGradient color.Color

	if fg == nil {
		gradient := lipgloss.Blend1D(
			height+width, // half of total number of border chars.
			Theme.DialogBorderGradientFromFg(),
			Theme.DialogBorderGradientToFg(),
		)

		// Duplicate the gradient (which is half size), and add it to the end of the gradient in reverse.
		rev := slices.Clone(gradient)
		slices.Reverse(rev)
		gradient = append(gradient, rev...)

		topGradient = gradient[0 : width-2]
		topRightCornerGradient = gradient[width-1]
		rightGradient = gradient[width : width+height-1]
		bottomRightCornerGradient = gradient[width+height]
		bottomGradient = gradient[width+height : width+height+width-2]
		bottomLeftCornerGradient = gradient[width+height+width]
		leftGradient = gradient[width+height+width+1:]
		topLeftCornerGradient = gradient[0]

		// bottom and left gradients are reversed because they are drawn in reverse order.
		slices.Reverse(bottomGradient)
		slices.Reverse(leftGradient)
	}

	wrapBrackets := func(text string) string {
		if text != "" {
			return fmt.Sprintf("%s%s%s",
				baseStyle.Render("["),
				text,
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
		// Add 2 to account for the padding border characters (1 on each side).
		remaining := max(0, width-lipgloss.Width(leftText)-lipgloss.Width(middleText)-lipgloss.Width(rightText)-2)
		leftBorderLen := max(0, (width/2)-lipgloss.Width(leftText)-(lipgloss.Width(middleText)/2)-1)
		rightBorderLen := max(0, remaining-leftBorderLen)

		// Build gradient border segments
		var leftBorderSegment strings.Builder
		for i := range leftBorderLen {
			var c color.Color
			if fg == nil {
				c = gradient[min(i, len(gradient)-1)]
			} else {
				c = fg
			}
			style := lipgloss.NewStyle().Foreground(c)
			leftBorderSegment.WriteString(style.Render(between))
		}

		var rightBorderSegment strings.Builder
		for i := range rightBorderLen {
			var c color.Color
			if fg == nil {
				c = gradient[min(leftBorderLen+i, len(gradient)-1)]
			} else {
				c = fg
			}
			style := lipgloss.NewStyle().Foreground(c)
			rightBorderSegment.WriteString(style.Render(between))
		}

		// Build padding border segments (1 character each)
		var leftPaddingBorder, rightPaddingBorder strings.Builder
		var leftPaddingColor, rightPaddingColor color.Color

		if fg == nil {
			// Left padding should use the first gradient color (index 0)
			// Right padding should use the gradient color at the position after all other elements
			leftPaddingColor = gradient[0]
			rightPaddingColor = gradient[min(leftBorderLen+lipgloss.Width(leftText)+lipgloss.Width(middleText)+lipgloss.Width(rightText)+rightBorderLen, len(gradient)-1)]
		} else {
			leftPaddingColor = fg
			rightPaddingColor = fg
		}

		leftPaddingStyle := lipgloss.NewStyle().Foreground(leftPaddingColor)
		rightPaddingStyle := lipgloss.NewStyle().Foreground(rightPaddingColor)
		leftPaddingBorder.WriteString(leftPaddingStyle.Render(between))
		rightPaddingBorder.WriteString(rightPaddingStyle.Render(between))

		var leftCornerStyle, rightCornerStyle lipgloss.Style
		if fg == nil {
			leftCornerStyle = lipgloss.NewStyle().Foreground(leftCornerGradient)
			rightCornerStyle = lipgloss.NewStyle().Foreground(rightCornerGradient)
		} else {
			leftCornerStyle = lipgloss.NewStyle().Foreground(fg)
			rightCornerStyle = lipgloss.NewStyle().Foreground(fg)
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
		var result []string

		for i, line := range lines {
			var leftColor, rightColor color.Color
			if fg == nil {
				leftColor = leftGradient[min(i, len(leftGradient)-1)]
				rightColor = rightGradient[min(i, len(rightGradient)-1)]
			} else {
				leftColor = fg
				rightColor = fg
			}

			leftBorderStyle := lipgloss.NewStyle().Foreground(leftColor)
			rightBorderStyle := lipgloss.NewStyle().Foreground(rightColor)

			// Add left and right borders to the line.
			borderedLine := leftBorderStyle.Render(border.Left) + line + rightBorderStyle.Render(border.Right)
			result = append(result, borderedLine)
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
			topGradient,
			topLeftCornerGradient,
			topRightCornerGradient,
		),
		buildVerticalBorders(content, leftGradient, rightGradient),
		buildHorizontalBorder(
			embeddedText[BottomLeftBorder],
			embeddedText[BottomMiddleBorder],
			embeddedText[BottomRightBorder],
			border.BottomLeft,
			border.Bottom,
			border.BottomRight,
			bottomGradient,
			bottomLeftCornerGradient,
			bottomRightCornerGradient,
		),
	}, "\n")
}
