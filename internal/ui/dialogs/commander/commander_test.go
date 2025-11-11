// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package commander

import (
	"os"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/pages/genericcode"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/testui"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true}) //nolint:errcheck
	os.Exit(v)
}

func newMockPageWithCommands(app types.AppState, commands []string, title, content, language string) types.Page {
	p := genericcode.New(app, title, content, language)
	p.Commands = commands
	return p
}

func newMockPageRef(app types.AppState, commands []string, title, content, language string) PageRef {
	return PageRef{
		Description: title,
		New: func() types.Page {
			return newMockPageWithCommands(app, commands, title, content, language)
		},
		Commands: commands,
	}
}

func newMockApp() types.AppState {
	app := state.NewMockAppState(api.NewMockClient(), nil)
	initial := newMockPageWithCommands(
		app,
		[]string{"current", "current-alias"},
		"Current Page",
		"current content",
		"text",
	)
	app.SetPage(state.NewPageState(initial))
	return app
}

func TestNew(t *testing.T) {
	t.Parallel()

	defaultWidth := 70
	defaultHeight := 10

	t.Run("basic-validation", func(t *testing.T) {
		t.Parallel()
		app := newMockApp()

		commander := New(app, Config{
			App: app,
			Pages: []PageRef{
				newMockPageRef(app, []string{"mock-page-1", "mock-page-1-alias"}, "Mock Page 1 Title", "mock content\nhere", "plaintext"),
				newMockPageRef(app, []string{"mock-page-2", "mock-page-2-alias"}, "Mock Page 2 Title", "mock content\nhere", "plaintext"),
				newMockPageRef(app, []string{"mock-page-3", "mock-page-3-alias"}, "Mock Page 3 Title", "mock content\nhere", "plaintext"),
				newMockPageRef(app, []string{"mock-page-4", "mock-page-4-alias"}, "Mock Page 4 Title", "mock content\nhere", "plaintext"),
			},
		})

		tm := testui.NewNonRootModel(t, commander, false, testui.WithTermSize(defaultWidth, defaultHeight))

		tm.ExpectContains(t,
			"mock-page-1",
			"mock-page-2",
			"mock-page-3",
			"mock-page-4",
		)
		tm.ExpectNotContains(t, "Current Page") // Since we didn't pass it in, and it has no commands.
		tm.ExpectViewSnapshot(t)

		go tm.Type("mock-page-1")
		tm.ExpectContains(t, "Mock Page 1 Title", "mock-page-1", "mock-page-1-alias")
		tm.ExpectNotContains(t, "mock-page-2", "mock-page-3", "mock-page-4")
		tm.ExpectViewSnapshot(t)

		for range 10 {
			tm.Send(tea.KeyPressMsg{Code: tea.KeyBackspace})
		}

		// We should see all of the commands again.
		tm.ExpectContains(t,
			"mock-page-1",
			"mock-page-2",
			"mock-page-3",
			"mock-page-4",
		)
		tm.ExpectViewSnapshot(t)

		tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter}) // Select first command.
		tm.WaitForFilterMessages(t, types.OpenPageMsg{})

		dialogMsgs := tm.WaitForFilterMessages(t, types.DialogMsg{})
		hasClose := false
		for _, msg := range dialogMsgs {
			switch msg := msg.(type) { //nolint:gocritic
			case types.DialogMsg:
				switch msg.Msg.(type) { //nolint:gocritic
				case types.CloseActiveDialogMsg:
					hasClose = true
				}
			}
		}
		if !hasClose {
			t.Fatalf("expected at least one CloseActiveDialogMsg, didn't find in %d messages", len(dialogMsgs))
		}
	})
}
