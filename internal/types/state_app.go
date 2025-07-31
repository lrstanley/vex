// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
)

type AppState interface {
	Page() PageState
	Dialog() DialogState
	Client() Client
	ShortHelp(focused FocusID, skip ...string) []key.Binding
	FullHelp(focused FocusID, skip ...string) [][]key.Binding
}

type AppQuitMsg struct{}

// AppQuit is sent when the user wants to quit the application. Don't use [tea.Quit],
// as different state may need to be cleaned up before quitting.
func AppQuit() tea.Cmd {
	return CmdMsg(AppQuitMsg{})
}

type FocusID string

const (
	FocusPage      FocusID = "page"
	FocusDialog    FocusID = "dialog"
	FocusStatusBar FocusID = "statusbar"
)

type AppFocusChangedMsg struct {
	ID FocusID
}

func FocusChange(id FocusID) tea.Cmd {
	return CmdMsg(AppFocusChangedMsg{ID: id})
}

type AppRequestPreviousFocusMsg struct{}

func RequestPreviousFocus() tea.Cmd {
	return CmdMsg(AppRequestPreviousFocusMsg{})
}

// AppFilterMsg is sent when the user provides a filter in the status bar.
type AppFilterMsg struct {
	UUID string
	Text string
}

func AppFilter(uuid, text string) tea.Cmd {
	return CmdMsg(AppFilterMsg{UUID: uuid, Text: text})
}

type AppFilterClearedMsg struct{}

func ClearAppFilter() tea.Cmd {
	return CmdMsg(AppFilterClearedMsg{})
}

func SetClipboard(content string) tea.Cmd {
	return tea.Batch(
		tea.SetClipboard(content),
		SendStatus("copied to clipboard", Info, 1*time.Second),
	)
}
