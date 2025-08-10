// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"image/color"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/bubbletint/chromatint/v2"
	tint "github.com/lrstanley/bubbletint/v2"
	"github.com/lrstanley/vex/internal/types"
)

// TODO: https://github.com/masaushi/accessory/pull/123
//go:generate go run github.com/lrstanley/accessory@latest -type ThemeConfig -receiver tc -lock mu -output theme.gen.go

var Theme = (&ThemeConfig{
	registry: tint.NewRegistry(
		tint.TintCga,
		tint.TintPencilDark,
		tint.TintDjango,
		tint.TintAfterglow,
		tint.TintSpacedust,
		tint.TintLabFox,
		tint.TintTokyoNightLight,
		tint.TintOneHalfLight,
		tint.TintGrape,
		tint.TintCyberPunk2077,
		tint.TintCyberdyne,
		tint.TintWryan,
		tint.TintUbuntu,
		tint.TintTomorrowNightBurns,
		tint.TintSolarizedDarkPatched,
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

	barBg                 color.Color `accessor:"getter"`
	barFg                 color.Color `accessor:"getter"`
	statusBarFilterTextFg color.Color `accessor:"getter"`
	statusBarFilterBg     color.Color `accessor:"getter"`
	statusBarFilterFg     color.Color `accessor:"getter"`
	statusBarAddrBg       color.Color `accessor:"getter"`
	statusBarAddrFg       color.Color `accessor:"getter"`
	statusBarUserFg       color.Color `accessor:"getter"`
	statusBarUserBg       color.Color `accessor:"getter"`
	statusBarTokenTTLFg   color.Color `accessor:"getter"`
	statusBarTokenTTLBg   color.Color `accessor:"getter"`
	statusBarLogoBg       color.Color `accessor:"getter"`
	statusBarLogoFg       color.Color `accessor:"getter"`

	shortHelpKeyFg color.Color `accessor:"getter"`

	dialogFg                   color.Color `accessor:"getter"`
	dialogBorderFg             color.Color `accessor:"getter"`
	dialogBorderGradientFromFg color.Color `accessor:"getter"`
	dialogBorderGradientToFg   color.Color `accessor:"getter"`

	titleFg     color.Color `accessor:"getter"`
	titleFromFg color.Color `accessor:"getter"`
	titleToFg   color.Color `accessor:"getter"`

	pageBorderFg       color.Color `accessor:"getter"`
	pageBorderFilterFg color.Color `accessor:"getter"`

	listItemFg         color.Color `accessor:"getter"`
	listItemSelectedFg color.Color `accessor:"getter"`
}

func (tc *ThemeConfig) Adapt(light, dark color.Color) color.Color {
	if tc.registry.Current().Dark {
		return dark
	}
	return light
}

// AdaptAuto adapts a color based on the current theme being light or dark. v is the
// float percentage to adjust the color by. If v is positive, dark will be lightened,
// and light will be darkened. If v is negative, dark will be darkened, and light will
// be lightened.
func (tc *ThemeConfig) AdaptAuto(c color.Color, v float64) color.Color {
	if tc.registry.Current().Dark {
		if v < 0 {
			return lipgloss.Darken(c, -v)
		}
		return lipgloss.Lighten(c, v)
	}
	if v < 0 {
		return lipgloss.Lighten(c, -v)
	}
	return lipgloss.Darken(c, v)
}

func (tc *ThemeConfig) set() *ThemeConfig {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	t := tc.registry.Current()

	tc.chroma = chromatint.StyleEntry(t, false)
	tc.fg = t.Fg

	white := tc.Adapt(lipgloss.Lighten(t.White, 0.2), lipgloss.Lighten(t.White, 0.2))

	statusFgLighten := 0.4
	statusBgDarken := 0.6

	tc.successFg = tc.Adapt(lipgloss.Lighten(t.BrightGreen, statusFgLighten), lipgloss.Lighten(t.BrightGreen, statusFgLighten))
	tc.successBg = tc.Adapt(lipgloss.Darken(t.BrightGreen, statusBgDarken), lipgloss.Darken(t.BrightGreen, statusBgDarken))
	tc.warningFg = tc.Adapt(lipgloss.Lighten(t.BrightYellow, statusFgLighten), lipgloss.Lighten(t.BrightYellow, statusFgLighten))
	tc.warningBg = tc.Adapt(lipgloss.Darken(t.BrightYellow, statusBgDarken), lipgloss.Darken(t.BrightYellow, statusBgDarken))
	tc.errorFg = tc.Adapt(lipgloss.Lighten(t.BrightRed, statusFgLighten), lipgloss.Lighten(t.BrightRed, statusFgLighten))
	tc.errorBg = tc.Adapt(lipgloss.Darken(t.BrightRed, statusBgDarken), lipgloss.Darken(t.BrightRed, statusBgDarken))
	tc.infoFg = tc.Adapt(lipgloss.Lighten(t.BrightBlue, statusFgLighten), lipgloss.Lighten(t.BrightBlue, statusFgLighten))
	tc.infoBg = tc.Adapt(lipgloss.Darken(t.BrightBlue, statusBgDarken), lipgloss.Darken(t.BrightBlue, statusBgDarken))

	tc.scrollbarThumbFg = tc.AdaptAuto(t.BrightBlue, 0.2)
	tc.scrollbarTrackFg = tc.AdaptAuto(t.Bg, 0.3)

	tc.barFg = tc.Adapt(t.Fg, t.Fg)
	tc.barBg = tc.Adapt(lipgloss.Lighten(t.Bg, 0.1), lipgloss.Darken(t.Bg, 0.2))
	tc.statusBarFilterTextFg = white
	tc.statusBarFilterBg = tc.infoBg
	tc.statusBarFilterFg = tc.infoFg
	tc.statusBarAddrFg = white
	tc.statusBarAddrBg = lipgloss.Darken(t.BrightBlue, 0.4)
	tc.statusBarUserFg = white
	tc.statusBarUserBg = lipgloss.Darken(t.BrightCyan, 0.4)
	tc.statusBarTokenTTLFg = white
	tc.statusBarTokenTTLBg = lipgloss.Darken(t.BrightYellow, 0.6)
	tc.statusBarLogoFg = white
	tc.statusBarLogoBg = tc.Adapt(t.Purple, lipgloss.Lighten(t.Bg, 0.2))

	tc.shortHelpKeyFg = tc.Adapt(lipgloss.Darken(t.BrightPurple, 0.3), lipgloss.Lighten(t.BrightPurple, 0.4))

	tc.dialogFg = white
	tc.dialogBorderFg = tc.Adapt(t.Purple, t.Purple)
	tc.dialogBorderGradientFromFg = tc.AdaptAuto(t.BrightPurple, 0.2)
	tc.dialogBorderGradientToFg = tc.AdaptAuto(t.BrightBlue, 0.2)

	tc.titleFg = tc.Adapt(lipgloss.Darken(t.BrightRed, 0.5), lipgloss.Lighten(t.BrightRed, 0.5))
	tc.titleFromFg = tc.Adapt(lipgloss.Darken(t.BrightPurple, 0.2), lipgloss.Lighten(t.BrightPurple, 0.2))
	tc.titleToFg = tc.Adapt(lipgloss.Darken(t.BrightBlue, 0.2), lipgloss.Lighten(t.BrightBlue, 0.2))

	tc.pageBorderFg = tc.Adapt(lipgloss.Darken(t.Purple, 0.2), lipgloss.Lighten(t.Purple, 0.2))
	tc.pageBorderFilterFg = tc.Adapt(lipgloss.Darken(t.BrightBlue, 0.3), lipgloss.Lighten(t.BrightBlue, 0.3))

	tc.listItemFg = tc.Adapt(lipgloss.Darken(t.BrightBlue, 0.6), lipgloss.Lighten(t.BrightBlue, 0.6))
	tc.listItemSelectedFg = tc.Adapt(lipgloss.Darken(t.BrightBlue, 0.2), lipgloss.Lighten(t.BrightBlue, 0.2))

	borderGradientCache.DeleteAll()
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
	return tea.Batch(
		types.SendStatus("Switched to "+tc.registry.Current().ID, types.Success, 1*time.Second),
		tc.updateThemeCmd(),
	)
}

func (tc *ThemeConfig) PreviousTint() tea.Cmd {
	tc.registry.PreviousTint()
	tc.set()
	return tea.Batch(
		types.SendStatus("Switched to "+tc.registry.Current().ID, types.Success, 1*time.Second),
		tc.updateThemeCmd(),
	)
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
