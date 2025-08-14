// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"slices"

	"github.com/charmbracelet/bubbles/v2/key"
)

var (
	// General purpose.

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
		key.WithHelp("d", "view details"),
	)
	KeyCopy = key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	)
	KeyHelp = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	)
	KeyQuit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)

	// Secret related.

	KeyToggleMask = key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "unmask"),
	)
	KeyToggleMaskAll = key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "unmask all"),
	)
	KeyRenderJSON = key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "view json"),
	)

	// Table related.

	KeysTable = []key.Binding{
		KeyTableLineUp,
		KeyTableLineDown,
		KeyTablePageUp,
		KeyTablePageDown,
		KeyTableGoToTop,
		KeyTableGoToBottom,
	}

	KeyTableLineUp = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)
	KeyTableLineDown = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	)
	KeyTableLineLeft = key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	)
	KeyTableLineRight = key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	)
	KeyTablePageUp = key.NewBinding(
		key.WithKeys("b", "pgup"),
		key.WithHelp("b/pgup", "page up"),
	)
	KeyTablePageDown = key.NewBinding(
		key.WithKeys("f", "pgdown"),
		key.WithHelp("f/pgdn", "page down"),
	)
	KeyTableGoToTop = key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
	)
	KeyTableGoToBottom = key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
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
