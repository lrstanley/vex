// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
)

// ClientCapabilities returns a string of the capabilities for a given path.
func ClientCapabilities(caps types.ClientCapabilities, path string) string {
	s := lipgloss.NewStyle().PaddingRight(1)

	var buf strings.Builder
	write := func(v string) {
		buf.WriteString(formatter.TruncReset(v))
	}

	isSingular := !strings.HasSuffix(path, "/")

	if len(caps) == 0 {
		write(s.Foreground(Theme.WarningFg()).Render("no perms"))
	}
	if caps.Contains(types.CapabilityRoot) {
		write(s.Foreground(Theme.ErrorFg()).Render(string(types.CapabilityRoot)))
		goto end
	}
	if caps.Contains(types.CapabilityDeny) {
		write(s.Foreground(Theme.ErrorFg()).Render(string(types.CapabilityDeny)))
		goto end
	}
	if caps.Contains(types.CapabilitySudo) {
		write(s.Foreground(Theme.ErrorFg()).Render(string(types.CapabilitySudo)))
		goto end
	}
	if caps.Contains(types.CapabilityWrite) {
		write(s.Foreground(Theme.SuccessFg()).Render(string(types.CapabilityWrite)))
	}
	if isSingular && caps.Contains(types.CapabilityRead) {
		write(s.Foreground(Theme.InfoFg()).Render(string(types.CapabilityRead)))
	}
	if !isSingular && caps.Contains(types.CapabilityList) {
		write(s.Foreground(Theme.InfoFg()).Render(string(types.CapabilityList)))
	}
	if isSingular && caps.Contains(types.CapabilityDelete) {
		write(s.Foreground(Theme.ErrorFg()).Render(string(types.CapabilityDelete)))
	}
	if !isSingular && caps.Contains(types.CapabilityCreate) {
		write(s.Foreground(Theme.SuccessFg()).Render(string(types.CapabilityCreate)))
	}
	if isSingular && caps.Contains(types.CapabilityUpdate) {
		write(s.Foreground(Theme.SuccessFg()).Render(string(types.CapabilityUpdate)))
	}
	if isSingular && caps.Contains(types.CapabilityPatch) {
		write(s.Foreground(Theme.SuccessFg()).Render(string(types.CapabilityPatch)))
	}
	if caps.Contains(types.CapabilitySubscribe) {
		write(s.Foreground(Theme.SuccessFg()).Render(string(types.CapabilitySubscribe)))
	}
	if caps.Contains(types.CapabilityRecover) {
		write(s.Foreground(Theme.SuccessFg()).Render(string(types.CapabilityRecover)))
	}

	if buf.Len() == 0 {
		write(s.Foreground(Theme.WarningFg()).Render("unknown"))
	}
end:

	return buf.String()
}
