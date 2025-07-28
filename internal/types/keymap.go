// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"slices"

	"github.com/charmbracelet/bubbles/v2/key"
)

var (
	// GlobalKeyBinds = []key.Binding{
	// 	KeyCommander,
	// 	KeyCommandBarFilter,
	// 	KeyHelp,
	// 	KeyQuit,
	// }

	KeyCommander = key.NewBinding(
		key.WithKeys(":"),
		key.WithHelp(":", "cmds"),
	)
	KeyFilter = key.NewBinding(
		key.WithKeys("/", "ctrl+f"),
		key.WithHelp("/", "filter"),
	)
	KeySelectItem = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select item"),
	)
	KeySelectItemAlt = key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "select item"),
	)
	KeyCancel = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	)
	KeyRefresh = key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "refresh"),
	)
	KeyDetails = key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "get details about resource"),
	)
	KeyCopy = key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy content"),
	)
	KeyHelp = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	)
	KeyQuit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
)

type KeyBindingGroup struct {
	Title    string
	Bindings [][]key.Binding
}

func KeyBindingContains(keys []key.Binding, against key.Binding) bool {
	in := against.Keys()
	var s []string
	for _, k := range keys {
		s = k.Keys()
		for _, v := range s {
			if slices.Contains(in, v) {
				return true
			}
		}
	}
	return false
}

func KeyBindingContainsFull(keys [][]key.Binding, against key.Binding) bool {
	for _, k := range keys {
		if KeyBindingContains(k, against) {
			return true
		}
	}
	return false
}
