// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"image/color"
	"sync"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2/colors"
	tint "github.com/lrstanley/bubbletint/v2"
	"github.com/lrstanley/vex/internal/types"
)

//go:generate go run github.com/masaushi/accessory@latest -type ThemeConfig -receiver tc -lock mu -output theme.gen.go

var Theme = (&ThemeConfig{
	registry: tint.NewRegistry(
		tint.TintPencilDark,
		tint.TintAfterglow,
		tint.TintSpacedust,
		tint.TintLabFox,
		// tint.DefaultTints()...,
	),
}).init(true)

type ThemeConfig struct {
	registry *tint.Registry
	mu       sync.RWMutex

	dark bool        `accessor:"getter"`
	fg   color.Color `accessor:"getter"`

	successFg color.Color `accessor:"getter"`
	successBg color.Color `accessor:"getter"`
	warningFg color.Color `accessor:"getter"`
	warningBg color.Color `accessor:"getter"`
	errorFg   color.Color `accessor:"getter"`
	errorBg   color.Color `accessor:"getter"`
	infoFg    color.Color `accessor:"getter"`
	infoBg    color.Color `accessor:"getter"`

	statusBarBg           color.Color `accessor:"getter"`
	statusBarFg           color.Color `accessor:"getter"`
	statusBarActivePageFg color.Color `accessor:"getter"`
	statusBarActivePageBg color.Color `accessor:"getter"`
	statusBarFilterTextFg color.Color `accessor:"getter"`
	statusBarFilterBg     color.Color `accessor:"getter"`
	statusBarFilterFg     color.Color `accessor:"getter"`
	statusBarAddrBg       color.Color `accessor:"getter"`
	statusBarAddrFg       color.Color `accessor:"getter"`
	statusBarLogoBg       color.Color `accessor:"getter"`
	statusBarLogoFg       color.Color `accessor:"getter"`

	shortHelpKeyFg color.Color `accessor:"getter"`

	dialogFg       color.Color `accessor:"getter"`
	dialogBorderFg color.Color `accessor:"getter"`

	dialogTitleFg     color.Color `accessor:"getter"`
	dialogTitleFromFg color.Color `accessor:"getter"`
	dialogTitleToFg   color.Color `accessor:"getter"`

	pageBorderFg color.Color `accessor:"getter"`
}

func (tc *ThemeConfig) adapt(light, dark color.Color) color.Color {
	if tc.dark {
		return dark
	}
	return light
}

func (tc *ThemeConfig) init(dark bool) *ThemeConfig {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	t := tc.registry.Current()

	tc.dark = dark

	tc.fg = tc.adapt(t.Fg, t.Fg)

	tc.successFg = tc.adapt(t.BrightGreen, t.BrightGreen)
	tc.successBg = tc.adapt(colors.Darken(t.BrightGreen, 60), colors.Darken(t.BrightGreen, 60))
	tc.warningFg = tc.adapt(t.BrightYellow, t.BrightYellow)
	tc.warningBg = tc.adapt(colors.Darken(t.BrightYellow, 60), colors.Darken(t.BrightYellow, 60))
	tc.errorFg = tc.adapt(colors.Lighten(t.BrightRed, 20), colors.Lighten(t.BrightRed, 20))
	tc.errorBg = tc.adapt(colors.Darken(t.BrightRed, 60), colors.Darken(t.BrightRed, 60))
	tc.infoFg = tc.adapt(colors.Lighten(t.BrightBlue, 20), colors.Lighten(t.BrightBlue, 20))
	tc.infoBg = tc.adapt(colors.Darken(t.BrightBlue, 60), colors.Darken(t.BrightBlue, 60))

	tc.statusBarFg = tc.adapt(t.Fg, t.Fg)
	tc.statusBarBg = tc.adapt(colors.Lighten(t.Bg, 10), colors.Lighten(t.Bg, 10))
	tc.statusBarActivePageFg = tc.adapt(colors.Lighten(t.BrightCyan, 40), colors.Lighten(t.BrightCyan, 40))
	tc.statusBarActivePageBg = tc.adapt(colors.Darken(t.BrightCyan, 40), colors.Darken(t.BrightCyan, 40))
	tc.statusBarFilterTextFg = tc.adapt(t.White, t.White)
	tc.statusBarFilterBg = tc.infoBg
	tc.statusBarFilterFg = tc.infoFg
	tc.statusBarAddrFg = tc.adapt(t.White, t.White)
	tc.statusBarAddrBg = tc.adapt(colors.Darken(t.BrightBlue, 40), colors.Darken(t.BrightBlue, 40))
	tc.statusBarLogoFg = tc.adapt(t.White, t.White)
	tc.statusBarLogoBg = tc.adapt(t.Purple, t.Purple)

	tc.shortHelpKeyFg = tc.adapt(colors.Lighten(t.BrightPurple, 20), colors.Lighten(t.BrightPurple, 20))

	tc.dialogFg = tc.adapt(t.White, t.White)
	tc.dialogBorderFg = tc.adapt(t.Purple, t.Purple)
	tc.dialogTitleFg = tc.adapt(colors.Lighten(t.BrightRed, 50), colors.Lighten(t.BrightRed, 50))
	tc.dialogTitleFromFg = tc.adapt(colors.Lighten(t.Blue, 20), colors.Lighten(t.Blue, 20))
	tc.dialogTitleToFg = tc.adapt(t.BrightPurple, t.BrightPurple)

	tc.pageBorderFg = tc.adapt(t.Purple, t.Purple)

	return tc
}

func (tc *ThemeConfig) ByStatus(status types.Status) (fg, bg color.Color) {
	switch status {
	case types.Success:
		return tc.successFg, tc.successBg
	case types.Info:
		return tc.infoFg, tc.infoBg
	case types.Warning:
		return tc.warningFg, tc.warningBg
	case types.Error:
		return tc.errorFg, tc.errorBg
	default:
		return tc.fg, nil
	}
}

func (tc *ThemeConfig) Update(dark bool) tea.Cmd {
	return types.CmdMsg(ThemeUpdatedMsg{Theme: tc.init(dark)})
}

func (tc *ThemeConfig) NextTint() tea.Cmd {
	return func() tea.Msg {
		tc.registry.NextTint()
		return ThemeUpdatedMsg{Theme: tc.init(tc.dark)}
	}
}

func (tc *ThemeConfig) PreviousTint() tea.Cmd {
	return func() tea.Msg {
		tc.registry.PreviousTint()
		return ThemeUpdatedMsg{Theme: tc.init(tc.dark)}
	}
}

type ThemeUpdatedMsg struct {
	Theme *ThemeConfig
}
