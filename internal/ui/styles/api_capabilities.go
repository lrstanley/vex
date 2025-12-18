// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
)

// ClientCapabilities returns a string of the capabilities for a given path.
func ClientCapabilities(baseStyle lipgloss.Style, caps types.ClientCapabilities, path string) lipgloss.Style {
	switch caps.Highest(path) {
	case types.CapabilityRoot, types.CapabilityDeny, types.CapabilitySudo, types.CapabilityDelete:
		return baseStyle.Foreground(Theme.ErrorFg())
	case types.CapabilityWrite, types.CapabilityCreate, types.CapabilityUpdate, types.CapabilityPatch, types.CapabilitySubscribe, types.CapabilityRecover:
		return baseStyle.Foreground(Theme.SuccessFg())
	case types.CapabilityRead, types.CapabilityList:
		return baseStyle.Foreground(Theme.InfoFg())
	default:
		return baseStyle.Foreground(Theme.WarningFg())
	}
}
