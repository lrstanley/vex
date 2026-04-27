// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package commander

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/pages/genericcode"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/steep"
)

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

		tm := steep.NewComponentHarness(t, commander, steep.WithInitialTermSize(defaultWidth, defaultHeight))

		tm.WaitContainsStrings(t, []string{
			"mock-page-1",
			"mock-page-2",
			"mock-page-3",
			"mock-page-4",
		})
		tm.WaitNotContainsString(t, "Current Page") // Since we didn't pass it in, and it has no commands.
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)

		tm.Type("mock-page-1")
		tm.WaitContainsStrings(t, []string{"Mock Page 1 Title", "mock-page-1", "mock-page-1-alias"})
		tm.WaitNotContainsStrings(t, []string{"mock-page-2", "mock-page-3", "mock-page-4"})
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)

		for range 10 {
			tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
		}

		// We should see all of the commands again.
		tm.WaitContainsStrings(t, []string{
			"mock-page-1",
			"mock-page-2",
			"mock-page-3",
			"mock-page-4",
		})
		tm.WaitSettleView(t).RequireSnapshotNoANSI(t)

		tm.Send(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})) // Select first command.
		steep.WaitMessage[types.OpenPageMsg](t, tm)

		dialogMsgs := steep.WaitMessages[types.DialogMsg](t, tm)
		hasClose := false
		for _, msg := range dialogMsgs {
			switch msg.Msg.(type) {
			case types.CloseActiveDialogMsg:
				hasClose = true
			}
		}
		if !hasClose {
			t.Fatalf("expected at least one CloseActiveDialogMsg, didn't find in %d messages", len(dialogMsgs))
		}
	})
}
