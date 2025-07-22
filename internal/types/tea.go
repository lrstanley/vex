// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// CmdMsg is a helper function to create a tea.Cmd that just returns the provided message.
func CmdMsg(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// MsgAfterDuration is a helper function to create a tea.Cmd that returns the provided
// message after the duration.
func MsgAfterDuration(msg tea.Msg, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(_ time.Time) tea.Msg {
		return msg
	})
}

// CmdAfterDuration is a helper function to invoke the tea.Cmd after the duration.
func CmdAfterDuration(cmd tea.Cmd, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(_ time.Time) tea.Msg {
		return cmd()
	})
}
