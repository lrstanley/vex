// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"github.com/charmbracelet/colorprofile"
)

var advancedColorProfiles = []colorprofile.Profile{
	colorprofile.ANSI256,
	colorprofile.TrueColor,
}

// refs:
//   - https://www.vertex42.com/ExcelTips/unicode-symbols.html
//   - https://www.amp-what.com/
//   - https://shapecatcher.com/
const (
	IconSeparator            = "•"
	IconEllipsis             = "…"
	IconOpenDottedCircle     = "◌"
	IconSemiFilledCircle     = "◎"
	IconClosedCircle         = "◉"
	IconFilledCircle         = "⏺"
	IconRefresh              = "⟳"
	IconTitleGradientDivider = "⫻"
	IconMaybeDanger          = "⁈"
	IconDanger               = "‼"
	IconUnknown              = "⁇"
	IconScrollbar            = "┃"
)

var (
	IconCaution      = iconFallback("⚠️", "⚠")
	IconFilter       = iconFallback("🔍", "⌕")
	IconUnderWeather = iconFallback("☔", "⛈")
	IconFlag         = iconFallback("🚩", "⚑")
)

func iconFallback(icon, fallback string) func() string {
	return func() string {
		if !Theme.SupportsAdvancedColors() {
			return fallback
		}
		return icon
	}
}
