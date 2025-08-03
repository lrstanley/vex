// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

func toHex(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

func chromaCompat(in lipgloss.Style) string {
	var s strings.Builder

	if v := in.GetForeground(); v != nil {
		s.WriteString(toHex(v) + " ")
	}
	//nolint:gocritic
	// if v := in.GetBackground(); v != nil {
	// 	s.WriteString("bg:" + toHex(v) + " ")
	// }
	if in.GetItalic() {
		s.WriteString("italic ")
	}
	if in.GetBold() {
		s.WriteString("bold ")
	}
	if in.GetUnderline() {
		s.WriteString("underline ")
	}
	return strings.TrimSpace(s.String())
}

func (tc *ThemeConfig) generateChromaStyle() chroma.StyleEntries {
	s := lipgloss.NewStyle()
	t := tc.registry.Current()
	return chroma.StyleEntries{ //nolint:exhaustive
		chroma.Other:                    chromaCompat(s.Foreground(t.Fg)),
		chroma.Error:                    chromaCompat(s.Foreground(t.Fg)),
		chroma.Background:               chromaCompat(s.Background(t.Bg)),
		chroma.Keyword:                  chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.KeywordConstant:          chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.KeywordDeclaration:       chromaCompat(s.Foreground(t.Cyan).Italic(true)),
		chroma.KeywordNamespace:         chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.KeywordPseudo:            chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.KeywordReserved:          chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.KeywordType:              chromaCompat(s.Foreground(t.Cyan)),
		chroma.Name:                     chromaCompat(s.Foreground(t.Fg)),
		chroma.NameAttribute:            chromaCompat(s.Foreground(t.Green)),
		chroma.NameBuiltin:              chromaCompat(s.Foreground(t.Cyan).Italic(true)),
		chroma.NameBuiltinPseudo:        chromaCompat(s.Foreground(t.Fg)),
		chroma.NameClass:                chromaCompat(s.Foreground(t.Green)),
		chroma.NameConstant:             chromaCompat(s.Foreground(t.Fg)),
		chroma.NameDecorator:            chromaCompat(s.Foreground(t.Fg)),
		chroma.NameEntity:               chromaCompat(s.Foreground(t.Fg)),
		chroma.NameException:            chromaCompat(s.Foreground(t.Fg)),
		chroma.NameFunction:             chromaCompat(s.Foreground(t.Green)),
		chroma.NameLabel:                chromaCompat(s.Foreground(t.Cyan).Italic(true)),
		chroma.NameNamespace:            chromaCompat(s.Foreground(t.Fg)),
		chroma.NameOther:                chromaCompat(s.Foreground(t.Fg)),
		chroma.NameTag:                  chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.NameVariable:             chromaCompat(s.Foreground(t.Cyan).Italic(true)),
		chroma.NameVariableClass:        chromaCompat(s.Foreground(t.Cyan).Italic(true)),
		chroma.NameVariableGlobal:       chromaCompat(s.Foreground(t.Cyan).Italic(true)),
		chroma.NameVariableInstance:     chromaCompat(s.Foreground(t.Cyan).Italic(true)),
		chroma.Literal:                  chromaCompat(s.Foreground(t.Fg)),
		chroma.LiteralDate:              chromaCompat(s.Foreground(t.Fg)),
		chroma.LiteralString:            chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringBacktick:    chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringChar:        chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringDoc:         chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringDouble:      chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringEscape:      chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringHeredoc:     chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringInterpol:    chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringOther:       chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringRegex:       chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringSingle:      chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralStringSymbol:      chromaCompat(s.Foreground(t.Yellow)),
		chroma.LiteralNumber:            chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4)))),
		chroma.LiteralNumberBin:         chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4)))),
		chroma.LiteralNumberFloat:       chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4)))),
		chroma.LiteralNumberHex:         chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4)))),
		chroma.LiteralNumberInteger:     chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4)))),
		chroma.LiteralNumberIntegerLong: chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4)))),
		chroma.LiteralNumberOct:         chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4)))),
		chroma.Operator:                 chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.OperatorWord:             chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.Punctuation:              chromaCompat(s.Foreground(t.BrightWhite)),
		chroma.Comment:                  chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4))).Faint(true)),
		chroma.CommentHashbang:          chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4))).Faint(true)),
		chroma.CommentMultiline:         chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4))).Faint(true)),
		chroma.CommentSingle:            chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4))).Faint(true)),
		chroma.CommentSpecial:           chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.Purple, 0.3), lipgloss.Lighten(t.Purple, 0.4))).Faint(true)),
		chroma.CommentPreproc:           chromaCompat(s.Foreground(tc.adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.2)))),
		chroma.Generic:                  chromaCompat(s.Foreground(t.Fg)),
		chroma.GenericDeleted:           chromaCompat(s.Foreground(t.Red)),
		chroma.GenericEmph:              chromaCompat(s.Foreground(t.Fg).Underline(true)),
		chroma.GenericError:             chromaCompat(s.Foreground(t.Fg)),
		chroma.GenericHeading:           chromaCompat(s.Foreground(t.Fg).Bold(true)),
		chroma.GenericInserted:          chromaCompat(s.Foreground(t.Green).Bold(true)),
		chroma.GenericOutput:            chromaCompat(s.Foreground(t.Fg).Faint(true)),
		chroma.GenericPrompt:            chromaCompat(s.Foreground(t.Fg)),
		chroma.GenericStrong:            chromaCompat(s.Foreground(t.Fg)),
		chroma.GenericSubheading:        chromaCompat(s.Foreground(t.Fg).Bold(true)),
		chroma.GenericTraceback:         chromaCompat(s.Foreground(t.Fg)),
		chroma.GenericUnderline:         chromaCompat(s.Underline(true)),
		chroma.Text:                     chromaCompat(s.Foreground(t.Fg)),
		chroma.TextWhitespace:           chromaCompat(s.Foreground(t.Fg)),
		chroma.TextPunctuation:          chromaCompat(s.Foreground(t.BrightWhite)),
	}
}
