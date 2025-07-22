// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ui

import (
	"github.com/lrstanley/vex/internal/types"
)

var _ types.AppState = &appState{} // Ensure that appState implements the AppState interface.

type appState struct {
	page   types.PageState
	dialog types.DialogState
	task   types.TaskState
	client types.Client
}

func (a *appState) Page() types.PageState {
	return a.page
}

func (a *appState) Dialog() types.DialogState {
	return a.dialog
}

func (a *appState) Task() types.TaskState {
	return a.task
}

func (a *appState) Client() types.Client {
	return a.client
}
