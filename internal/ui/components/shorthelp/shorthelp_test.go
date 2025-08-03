// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package shorthelp

import (
	"os"
	"testing"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/ui/testui"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true}) //nolint:errcheck
	os.Exit(v)
}

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-help", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetMaxWidth(testui.DefaultTermWidth)
		m.SetKeyBinds(
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "help")),
		)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})

	t.Run("no-keybinds", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetMaxWidth(testui.DefaultTermWidth)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})

	t.Run("disabled-keybind", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetMaxWidth(testui.DefaultTermWidth)
		kb := key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
		kb.SetEnabled(false)
		m.SetKeyBinds(kb)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.SetMaxWidth(0)
		m.SetKeyBinds(
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})

	t.Run("max-width", func(t *testing.T) {
		t.Parallel()
		m := New()
		m.Height = testui.DefaultTermHeight
		m.Width = testui.DefaultTermWidth
		m.SetMaxWidth(20)
		m.SetKeyBinds(
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "help")),
		)
		tm := testui.NewNonRootModel(t, m, false)
		tm.ExpectViewSnapshot(t)
	})
}
