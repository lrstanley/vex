// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"image/color"
	"sync"

	"github.com/alecthomas/chroma/v2"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
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

	chroma chroma.StyleEntries
	fg     color.Color `accessor:"getter"`

	successFg color.Color `accessor:"getter"`
	successBg color.Color `accessor:"getter"`
	warningFg color.Color `accessor:"getter"`
	warningBg color.Color `accessor:"getter"`
	errorFg   color.Color `accessor:"getter"`
	errorBg   color.Color `accessor:"getter"`
	infoFg    color.Color `accessor:"getter"`
	infoBg    color.Color `accessor:"getter"`

	scrollbarThumbFg color.Color `accessor:"getter"`
	scrollbarTrackFg color.Color `accessor:"getter"`

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

	dialogFg                   color.Color `accessor:"getter"`
	dialogBorderFg             color.Color `accessor:"getter"`
	dialogBorderGradientFromFg color.Color `accessor:"getter"`
	dialogBorderGradientToFg   color.Color `accessor:"getter"`

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

	tc.chroma = tc.generateChromaStyle()
	tc.fg = tc.adapt(t.Fg, t.Fg)

	white := tc.adapt(lipgloss.Lighten(t.White, 0.2), lipgloss.Lighten(t.White, 0.2))

	statusFgLighten := 0.4
	statusBgDarken := 0.6

	tc.successFg = tc.adapt(lipgloss.Lighten(t.BrightGreen, statusFgLighten), lipgloss.Lighten(t.BrightGreen, statusFgLighten))
	tc.successBg = tc.adapt(lipgloss.Darken(t.BrightGreen, statusBgDarken), lipgloss.Darken(t.BrightGreen, statusBgDarken))
	tc.warningFg = tc.adapt(lipgloss.Lighten(t.BrightYellow, statusFgLighten), lipgloss.Lighten(t.BrightYellow, statusFgLighten))
	tc.warningBg = tc.adapt(lipgloss.Darken(t.BrightYellow, statusBgDarken), lipgloss.Darken(t.BrightYellow, statusBgDarken))
	tc.errorFg = tc.adapt(lipgloss.Lighten(t.BrightRed, statusFgLighten), lipgloss.Lighten(t.BrightRed, statusFgLighten))
	tc.errorBg = tc.adapt(lipgloss.Darken(t.BrightRed, statusBgDarken), lipgloss.Darken(t.BrightRed, statusBgDarken))
	tc.infoFg = tc.adapt(lipgloss.Lighten(t.BrightBlue, statusFgLighten), lipgloss.Lighten(t.BrightBlue, statusFgLighten))
	tc.infoBg = tc.adapt(lipgloss.Darken(t.BrightBlue, statusBgDarken), lipgloss.Darken(t.BrightBlue, statusBgDarken))

	tc.scrollbarThumbFg = tc.adapt(lipgloss.Darken(t.BrightBlue, 0.2), lipgloss.Lighten(t.BrightBlue, 0.2))
	tc.scrollbarTrackFg = tc.adapt(lipgloss.Darken(t.Bg, 0.3), lipgloss.Lighten(t.Bg, 0.3))

	tc.statusBarFg = tc.adapt(t.Fg, t.Fg)
	tc.statusBarBg = tc.adapt(lipgloss.Lighten(t.Bg, 0.1), lipgloss.Darken(t.Bg, 0.2))
	tc.statusBarActivePageFg = tc.adapt(lipgloss.Lighten(t.BrightCyan, 0.4), lipgloss.Lighten(t.BrightCyan, 0.4))
	tc.statusBarActivePageBg = tc.adapt(lipgloss.Darken(t.BrightCyan, 0.4), lipgloss.Darken(t.BrightCyan, 0.4))
	tc.statusBarFilterTextFg = white
	tc.statusBarFilterBg = tc.infoBg
	tc.statusBarFilterFg = tc.infoFg
	tc.statusBarAddrFg = white
	tc.statusBarAddrBg = tc.adapt(lipgloss.Darken(t.BrightBlue, 0.4), lipgloss.Darken(t.BrightBlue, 0.4))
	tc.statusBarLogoFg = white
	tc.statusBarLogoBg = tc.adapt(t.Purple, lipgloss.Lighten(t.Bg, 0.2))

	tc.shortHelpKeyFg = tc.adapt(lipgloss.Darken(t.BrightPurple, 0.2), lipgloss.Lighten(t.BrightPurple, 0.4))

	tc.dialogFg = white
	tc.dialogBorderFg = tc.adapt(t.Purple, t.Purple)
	tc.dialogBorderGradientFromFg = tc.adapt(lipgloss.Darken(t.BrightPurple, 0.2), lipgloss.Lighten(t.BrightPurple, 0.2))
	tc.dialogBorderGradientToFg = tc.adapt(lipgloss.Darken(t.BrightBlue, 0.2), lipgloss.Lighten(t.BrightBlue, 0.2))

	tc.dialogTitleFg = tc.adapt(lipgloss.Darken(t.BrightRed, 0.5), lipgloss.Lighten(t.BrightRed, 0.5))
	tc.dialogTitleFromFg = tc.adapt(lipgloss.Darken(t.BrightPurple, 0.2), lipgloss.Lighten(t.BrightPurple, 0.2))
	tc.dialogTitleToFg = tc.adapt(lipgloss.Darken(t.BrightBlue, 0.2), lipgloss.Lighten(t.BrightBlue, 0.2))

	tc.pageBorderFg = tc.adapt(t.Purple, t.Purple)
	tc.pageBorderFilterFg = tc.adapt(lipgloss.Darken(t.BrightBlue, 0.3), lipgloss.Lighten(t.BrightBlue, 0.3))

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

func (tc *ThemeConfig) Chroma() chroma.StyleEntries {
	if tc == nil {
		return nil
	}
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.chroma
}

type ThemeUpdatedMsg struct{}
