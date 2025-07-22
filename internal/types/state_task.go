// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"context"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type TaskState interface {
	AddTask(id string, metadata map[string]any, cancel context.CancelFunc) tea.Cmd
	Run(id string, metadata map[string]any, fn func(context.Context) tea.Cmd) tea.Cmd
	CancelTask(id string) tea.Cmd
	CancelTasksByFilter(filters map[string]any) tea.Cmd
	GetTask(id string) *Task
	GetAllTasks() map[string]*Task
	GetTaskCount() int
	CancelAll() tea.Cmd
}

// Task holds information about a background task that is running, and allows
// for cancellation through an external event stream (e.g. tea messages).
type Task struct {
	ID       string
	Metadata map[string]any
	Cancel   context.CancelFunc
}

// TaskMsg wraps all task related messages.
type TaskMsg struct {
	Msg any
}

// TaskStartedMsg is emitted when a task is started. Should always be wrapped in a TaskMsg.
type TaskStartedMsg struct {
	ID       string
	Metadata map[string]any
}

// TaskCancelledMsg is emitted when a task is cancelled. Should always be wrapped in a TaskMsg.
type TaskCancelledMsg struct {
	ID       string
	Metadata map[string]any
}
