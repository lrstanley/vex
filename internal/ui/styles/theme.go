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
		tint.TintDjango,
		tint.TintAfterglow,
		tint.TintSpacedust,
		tint.TintLabFox,
		tint.TintTokyoNightLight,
		// tint.DefaultTints()...,
	),
}).set()

type ThemeConfig struct {
	registry *tint.Registry
	mu       sync.RWMutex

	fg color.Color `accessor:"getter"`

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

	pageBorderFg       color.Color `accessor:"getter"`
	pageBorderFilterFg color.Color `accessor:"getter"`
}

func (tc *ThemeConfig) adapt(light, dark color.Color) color.Color {
	if tc.registry.Current().Dark {
		return dark
	}
	return light
}

func (tc *ThemeConfig) set() *ThemeConfig {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	t := tc.registry.Current()

	tc.fg = tc.adapt(t.Fg, t.Fg)

	white := tc.adapt(colors.Lighten(t.White, 20), colors.Lighten(t.White, 20))

	statusFgLighten := 40
	statusBgDarken := 60

	tc.successFg = tc.adapt(colors.Lighten(t.BrightGreen, statusFgLighten), colors.Lighten(t.BrightGreen, statusFgLighten))
	tc.successBg = tc.adapt(colors.Darken(t.BrightGreen, statusBgDarken), colors.Darken(t.BrightGreen, statusBgDarken))
	tc.warningFg = tc.adapt(colors.Lighten(t.BrightYellow, statusFgLighten), colors.Lighten(t.BrightYellow, statusFgLighten))
	tc.warningBg = tc.adapt(colors.Darken(t.BrightYellow, statusBgDarken), colors.Darken(t.BrightYellow, statusBgDarken))
	tc.errorFg = tc.adapt(colors.Lighten(t.BrightRed, statusFgLighten), colors.Lighten(t.BrightRed, statusFgLighten))
	tc.errorBg = tc.adapt(colors.Darken(t.BrightRed, statusBgDarken), colors.Darken(t.BrightRed, statusBgDarken))
	tc.infoFg = tc.adapt(colors.Lighten(t.BrightBlue, statusFgLighten), colors.Lighten(t.BrightBlue, statusFgLighten))
	tc.infoBg = tc.adapt(colors.Darken(t.BrightBlue, statusBgDarken), colors.Darken(t.BrightBlue, statusBgDarken))

	tc.statusBarFg = tc.adapt(t.Fg, t.Fg)
	tc.statusBarBg = tc.adapt(colors.Lighten(t.Bg, 20), colors.Darken(t.Bg, 20))
	tc.statusBarActivePageFg = tc.adapt(colors.Lighten(t.BrightCyan, 40), colors.Lighten(t.BrightCyan, 40))
	tc.statusBarActivePageBg = tc.adapt(colors.Darken(t.BrightCyan, 40), colors.Darken(t.BrightCyan, 40))
	tc.statusBarFilterTextFg = white
	tc.statusBarFilterBg = tc.infoBg
	tc.statusBarFilterFg = tc.infoFg
	tc.statusBarAddrFg = white
	tc.statusBarAddrBg = tc.adapt(colors.Darken(t.BrightBlue, 40), colors.Darken(t.BrightBlue, 40))
	tc.statusBarLogoFg = white
	tc.statusBarLogoBg = tc.adapt(t.Purple, colors.Lighten(t.Bg, 20))

	tc.shortHelpKeyFg = tc.adapt(colors.Lighten(t.BrightPurple, 20), colors.Lighten(t.BrightPurple, 20))

	tc.dialogFg = white
	tc.dialogBorderFg = tc.adapt(t.Purple, t.Purple)
	tc.dialogTitleFg = tc.adapt(colors.Darken(t.BrightRed, 50), colors.Lighten(t.BrightRed, 50))
	tc.dialogTitleFromFg = tc.adapt(colors.Lighten(t.Blue, 20), colors.Lighten(t.Blue, 20))
	tc.dialogTitleToFg = tc.adapt(t.BrightPurple, t.BrightPurple)

	tc.pageBorderFg = tc.adapt(t.Purple, t.Purple)
	tc.pageBorderFilterFg = tc.adapt(colors.Darken(t.BrightBlue, 30), colors.Lighten(t.BrightBlue, 30))

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

func (tc *ThemeConfig) Init() tea.Cmd {
	tc.set()
	return tea.Sequence(
		types.CmdMsg(ThemeUpdatedMsg{}),
		tea.SetBackgroundColor(tc.registry.Current().Bg),
	)
}

func (tc *ThemeConfig) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case tea.BackgroundColorMsg:
		// TODO: if user hasn't explicitly configured a tint, we should switch to
		// one which is the same as the current background color. We should also
		// make setting of the background color optional.
		tc.set()
		return tc.updateThemeCmd()
	}
	return nil
}

func (tc *ThemeConfig) NextTint() tea.Cmd {
	tc.registry.NextTint()
	tc.set()
	return tc.updateThemeCmd()
}

func (tc *ThemeConfig) PreviousTint() tea.Cmd {
	tc.registry.PreviousTint()
	tc.set()
	return tc.updateThemeCmd()
}

func (tc *ThemeConfig) updateThemeCmd() tea.Cmd {
	return tea.Batch(
		types.CmdMsg(ThemeUpdatedMsg{}),
		tea.SetBackgroundColor(tc.registry.Current().Bg),
	)
}

type ThemeUpdatedMsg struct{}
