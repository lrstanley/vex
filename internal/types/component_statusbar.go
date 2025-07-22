// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Status represents the state of a temporary message that we want to display to the user.
type Status string

const (
	Success Status = "success"
	Info    Status = "info"
	Warning Status = "warning"
	Error   Status = "error"
)

// StatusMsg wraps all status related messages.
type StatusMsg struct {
	Msg any
}

// StatusTextMsg is a message to display a status text message. Should always be
// wrapped in a StatusMsg.
type StatusTextMsg struct {
	Status   Status
	Text     string
	Duration time.Duration

	ID int64 // ID of the status text message. Automatically set by [SendStatus*] functions.
}

// SendStatus is a message to display a temporary status text message.
func SendStatus(text string, status Status, duration time.Duration) tea.Cmd {
	id := time.Now().UnixNano()
	cmds := []tea.Cmd{
		CmdMsg(StatusMsg{Msg: StatusTextMsg{
			Status:   status,
			Text:     text,
			Duration: duration,
			ID:       id,
		}}),
	}

	if duration > 0 {
		cmds = append(cmds, CmdAfterDuration(ClearStatusText(id), duration))
	}

	return tea.Batch(cmds...)
}

// ClearStatusTextMsg is a message to clear a status text message. Should always be
// wrapped in a StatusMsg.
type ClearStatusTextMsg struct {
	ID int64
}

// ClearStatusText is a helper function to clear a status text message.
func ClearStatusText(id int64) tea.Cmd {
	return CmdMsg(StatusMsg{Msg: ClearStatusTextMsg{ID: id}})
}

// StatusOperationMsg is a message to add an operation to the statusbar. Should
// always be wrapped in a StatusMsg.
type StatusOperationMsg struct {
	ID   int64
	Text string
}

// AddStatusOperation is a helper function to add a status operation message.
func AddStatusOperation(id int64, text string) tea.Cmd {
	return CmdMsg(StatusMsg{Msg: StatusOperationMsg{ID: id, Text: text}})
}

// ClearStatusOperationMsg is a message to clear a status operation message. Should
// always be wrapped in a StatusMsg.
type ClearStatusOperationMsg struct {
	ID int64
}

// ClearStatusOperation is a helper function to clear a status operation message.
func ClearStatusOperation(id int64) tea.Cmd {
	return CmdMsg(StatusMsg{Msg: ClearStatusOperationMsg{ID: id}})
}
