// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

type AppState interface {
	Page() PageState
	Dialog() DialogState
	Task() TaskState
	Client() Client
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
