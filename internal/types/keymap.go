// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"slices"

	"charm.land/bubbles/v2/key"
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
	KeyUp = key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)
	KeyDown = key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	)
	KeyLeft = key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "left"),
	)
	KeyRight = key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "right"),
	)
	KeyPageUp = key.NewBinding(
		key.WithKeys("b", "pgup"),
		key.WithHelp("b/pgup", "page up"),
	)
	KeyPageDown = key.NewBinding(
		key.WithKeys("f", "pgdown"),
		key.WithHelp("f/pgdn", "page down"),
	)
	KeyGoToTop = key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
	)
	KeyGoToBottom = key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
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
	KeyTabForward = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "tab forward"),
	)
	KeyTabBackward = key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "tab backward"),
	)
	KeyHelp = key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	)
	KeyQuit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	KeyDelete = key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "delete"),
	)
	KeyDestroy = key.NewBinding(
		key.WithKeys("ctrl+k"),
		key.WithHelp("ctrl+k", "destroy"),
	)
	KeyOpenEditor = key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "open in editor"),
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
	KeyToggleDelete = key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "toggle delete"),
	)
	KeyRenderJSON = key.NewBinding(
		key.WithKeys("z"),
		key.WithHelp("z", "view json"),
	)
	KeyListRecursive = key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "list secrets recursively"),
	)

	// Table related.

	KeysTable = []key.Binding{
		KeyUp,
		KeyDown,
		KeyLeft,
		KeyRight,
		KeyPageUp,
		KeyPageDown,
		KeyGoToTop,
		KeyGoToBottom,
	}
)

// OverrideHelp overrides the help text for a key binding, returning a new key
// binding with the same keys and the new help text.
func OverrideHelp(b key.Binding, help string) key.Binding {
	return key.NewBinding(
		key.WithKeys(b.Keys()...),
		key.WithHelp(b.Help().Key, help),
	)
}

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
