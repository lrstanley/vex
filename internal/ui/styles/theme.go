// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"image/color"
	"slices"
	"sync"
	"time"

	"github.com/alecthomas/chroma/v2"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
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
	profile: colorprofile.ANSI,
	chroma:  chromastyles.Fallback,
}).set()

type ThemeConfig struct {
	registry *tint.Registry
	mu       sync.RWMutex

	profile                colorprofile.Profile
	supportsAdvancedColors bool `accessor:"getter"`

	chroma *chroma.Style
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

func (tc *ThemeConfig) adapt(light, dark color.Color) color.Color {
	if tc.registry.Current().Dark || tc.supportsAdvancedColors {
		return dark
	}
	return light
}

// adaptAuto adapts a color based on the current theme being light or dark. v is the
// float percentage to adjust the color by. If v is positive, dark will be lightened,
// and light will be darkened. If v is negative, dark will be darkened, and light will
// be lightened.
func (tc *ThemeConfig) adaptAuto(c color.Color, v float64) color.Color {
	if tc.registry.Current().Dark {
		if v < 0 {
			return tc.darken(c, -v)
		}
		return tc.lighten(c, v)
	}
	if v < 0 {
		return tc.lighten(c, -v)
	}
	return tc.darken(c, v)
}

func (tc *ThemeConfig) darken(c color.Color, v float64) color.Color {
	return lipgloss.Darken(c, v)
}

func (tc *ThemeConfig) lighten(c color.Color, v float64) color.Color {
	return lipgloss.Lighten(c, v)
}

func (tc *ThemeConfig) useFallback() {
	tc.fg = lipgloss.White

	tc.successFg = lipgloss.White
	tc.successBg = lipgloss.Green
	tc.warningFg = lipgloss.White
	tc.warningBg = lipgloss.Yellow
	tc.errorFg = lipgloss.White
	tc.errorBg = lipgloss.Red
	tc.infoFg = lipgloss.White
	tc.infoBg = lipgloss.Blue

	tc.scrollbarThumbFg = lipgloss.BrightBlue
	tc.scrollbarTrackFg = lipgloss.Black

	tc.barFg = lipgloss.White
	tc.barBg = lipgloss.Black
	tc.statusBarFilterTextFg = lipgloss.White
	tc.statusBarFilterBg = lipgloss.Blue
	tc.statusBarFilterFg = lipgloss.White
	tc.statusBarAddrFg = lipgloss.White
	tc.statusBarAddrBg = lipgloss.Blue
	tc.statusBarUserFg = lipgloss.White
	tc.statusBarUserBg = lipgloss.Blue
	tc.statusBarTokenTTLFg = lipgloss.White
	tc.statusBarTokenTTLBg = lipgloss.Yellow
	tc.statusBarLogoFg = lipgloss.White
	tc.statusBarLogoBg = lipgloss.Magenta

	tc.shortHelpKeyFg = lipgloss.BrightMagenta

	tc.dialogFg = lipgloss.White
	tc.dialogBorderFg = lipgloss.BrightMagenta
	tc.dialogBorderGradientFromFg = lipgloss.BrightMagenta
	tc.dialogBorderGradientToFg = lipgloss.BrightMagenta

	tc.titleFg = lipgloss.BrightRed
	tc.titleFromFg = lipgloss.BrightMagenta
	tc.titleToFg = lipgloss.BrightMagenta

	tc.pageBorderFg = lipgloss.Magenta
	tc.pageBorderFilterFg = lipgloss.BrightBlue

	tc.listItemFg = lipgloss.BrightBlue
	tc.listItemSelectedFg = lipgloss.BrightBlue
}

func (tc *ThemeConfig) set() *ThemeConfig {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	t := tc.registry.Current()

	if !tc.supportsAdvancedColors {
		t = tint.TintITerm2Default
		tc.useFallback()
		borderGradientCache.DeleteAll()
		return tc
	}

	if cs, _ := chroma.NewStyle("vex", chromatint.StyleEntry(t, false)); cs != nil {
		tc.chroma = cs
	}

	tc.fg = t.Fg
	white := tc.adapt(tc.lighten(t.White, 0.2), tc.lighten(t.White, 0.2))

	statusFgLighten := 0.4
	statusBgDarken := 0.6

	tc.successFg = tc.adapt(tc.lighten(t.BrightGreen, statusFgLighten/2), tc.lighten(t.BrightGreen, statusFgLighten/2))
	tc.successBg = tc.adapt(tc.darken(t.BrightGreen, statusBgDarken), tc.darken(t.BrightGreen, statusBgDarken))
	tc.warningFg = tc.adapt(tc.lighten(t.BrightYellow, statusFgLighten/2), tc.lighten(t.BrightYellow, statusFgLighten/2))
	tc.warningBg = tc.adapt(tc.darken(t.BrightYellow, statusBgDarken), tc.darken(t.BrightYellow, statusBgDarken))
	tc.errorFg = tc.adapt(tc.lighten(t.BrightRed, statusFgLighten/2), tc.lighten(t.BrightRed, statusFgLighten/2))
	tc.errorBg = tc.adapt(tc.darken(t.BrightRed, statusBgDarken), tc.darken(t.BrightRed, statusBgDarken))
	tc.infoFg = tc.adapt(tc.lighten(t.BrightBlue, statusFgLighten/2), tc.lighten(t.BrightBlue, statusFgLighten/2))
	tc.infoBg = tc.adapt(tc.darken(t.BrightBlue, statusBgDarken), tc.darken(t.BrightBlue, statusBgDarken))

	tc.scrollbarThumbFg = tc.adaptAuto(t.BrightBlue, 0.2)
	tc.scrollbarTrackFg = tc.adaptAuto(t.Bg, 0.3)

	tc.barFg = tc.adapt(t.Fg, t.Fg)
	tc.barBg = tc.adapt(tc.lighten(t.Bg, 0.1), tc.darken(t.Bg, 0.2))
	tc.statusBarFilterTextFg = white
	tc.statusBarFilterBg = tc.infoBg
	tc.statusBarFilterFg = tc.infoFg
	tc.statusBarAddrFg = white
	tc.statusBarAddrBg = tc.darken(t.BrightBlue, 0.4)
	tc.statusBarUserFg = white
	tc.statusBarUserBg = tc.darken(t.BrightCyan, 0.4)
	tc.statusBarTokenTTLFg = white
	tc.statusBarTokenTTLBg = tc.darken(t.BrightYellow, 0.6)
	tc.statusBarLogoFg = white
	tc.statusBarLogoBg = tc.adapt(t.Purple, tc.lighten(t.Bg, 0.2))

	tc.shortHelpKeyFg = tc.adapt(tc.darken(t.BrightPurple, 0.3), tc.lighten(t.BrightPurple, 0.4))

	tc.dialogFg = white
	tc.dialogBorderFg = tc.adapt(t.Purple, t.Purple)
	tc.dialogBorderGradientFromFg = tc.adaptAuto(t.BrightPurple, 0.2)
	tc.dialogBorderGradientToFg = tc.adaptAuto(t.BrightBlue, 0.2)

	tc.titleFg = tc.adapt(tc.darken(t.BrightRed, 0.5), tc.lighten(t.BrightRed, 0.5))
	tc.titleFromFg = tc.adapt(tc.darken(t.BrightPurple, 0.2), tc.lighten(t.BrightPurple, 0.2))
	tc.titleToFg = tc.adapt(tc.darken(t.BrightBlue, 0.2), tc.lighten(t.BrightBlue, 0.2))

	tc.pageBorderFg = tc.adapt(tc.darken(t.Purple, 0.2), tc.lighten(t.Purple, 0.2))
	tc.pageBorderFilterFg = tc.adapt(tc.darken(t.BrightBlue, 0.3), tc.lighten(t.BrightBlue, 0.3))

	tc.listItemFg = tc.adapt(tc.darken(t.BrightBlue, 0.6), tc.lighten(t.BrightBlue, 0.6))
	tc.listItemSelectedFg = tc.adapt(tc.darken(t.BrightBlue, 0.2), tc.lighten(t.BrightBlue, 0.2))

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
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// TODO: if user hasn't explicitly configured a tint, we should switch to
		// one which is the same as the current background color. We should also
		// make setting of the background color optional.
		tc.set()
		return tc.updateThemeCmd()
	case tea.ColorProfileMsg:
		tc.mu.Lock()
		tc.profile = msg.Profile
		tc.supportsAdvancedColors = slices.Contains(advancedColorProfiles, msg.Profile)
		tc.mu.Unlock()
		tc.set()
		return tc.updateThemeCmd()
	}
	return nil
}

func (tc *ThemeConfig) NextTint() tea.Cmd {
	if !tc.SupportsAdvancedColors() {
		return nil
	}
	tc.registry.NextTint()
	tc.set()
	return tea.Batch(
		types.SendStatus("Switched to "+tc.registry.Current().ID, types.Success, 1*time.Second),
		tc.updateThemeCmd(),
	)
}

func (tc *ThemeConfig) PreviousTint() tea.Cmd {
	if !tc.SupportsAdvancedColors() {
		return nil
	}
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

func (tc *ThemeConfig) Chroma() *chroma.Style {
	if tc == nil {
		return nil
	}
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.chroma
}

type ThemeUpdatedMsg struct{}
