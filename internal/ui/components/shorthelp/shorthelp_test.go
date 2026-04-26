// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package shorthelp

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/lrstanley/x/charm/steep"
)

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-help", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetMaxWidth(steep.DefaultTermWidth)
		m.SetKeyBinds(
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "help")),
		)
		tm := steep.NewViewModel(t, m)
		tm.WaitContainsStrings(t, []string{"quit", "help"})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("no-keybinds", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetMaxWidth(steep.DefaultTermWidth)
		tm := steep.NewViewModel(t, m)
		tm.WaitSettleMessages(t).RequireSnapshotNoANSI(t)
	})

	t.Run("disabled-keybind", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetMaxWidth(steep.DefaultTermWidth)
		kb := key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
		kb.SetEnabled(false)
		m.SetKeyBinds(kb)
		tm := steep.NewViewModel(t, m)
		tm.WaitSettleMessages(t).RequireSnapshotNoANSI(t)
	})

	t.Run("max-width", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.Height = steep.DefaultTermHeight
		m.Width = steep.DefaultTermWidth
		m.SetMaxWidth(20)
		m.SetKeyBinds(
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "help")),
		)
		tm := steep.NewViewModel(t, m)
		tm.WaitContainsStrings(t, []string{"quit", "help"})
		tm.RequireSnapshotNoANSI(t)
	})
}
