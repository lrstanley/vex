// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ui

import (
	"github.com/charmbracelet/bubbles/v2/key"
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

func (a *appState) ShortHelp(focused types.FocusID, skip ...string) []key.Binding {
	keys := a.page.Get().ShortHelp()

	if focused == types.FocusDialog {
		if dialog := a.dialog.GetWithSkip(skip...); dialog != nil {
			keys = append(dialog.ShortHelp(), keys...)
		}
	}

	var prepended []key.Binding

	if a.page.Get().GetSupportFiltering() && !types.KeyBindingContains(keys, types.KeyFilter) {
		prepended = append(prepended, types.KeyFilter)
	}

	if !types.KeyBindingContains(keys, types.KeyCommander) {
		prepended = append(prepended, types.KeyCommander)
	}

	if !types.KeyBindingContains(keys, types.KeyHelp) {
		prepended = append(prepended, types.KeyHelp)
	}

	if !types.KeyBindingContains(keys, types.KeyQuit) {
		keys = append(keys, types.KeyQuit)
	}

	return append(prepended, keys...)
}

func (a *appState) FullHelp(focused types.FocusID, skip ...string) [][]key.Binding {
	keys := a.page.Get().FullHelp()

	if focused == types.FocusDialog {
		if dialog := a.dialog.GetWithSkip(skip...); dialog != nil {
			keys = append(dialog.FullHelp(), keys...)
		}
	}

	var prepend, appended []key.Binding

	if a.page.HasParent() && !types.KeyBindingContainsFull(keys, types.KeyCancel) {
		prepend = append(prepend, types.KeyCancel)
	}

	if a.page.Get().GetRefreshInterval() > 0 && !types.KeyBindingContainsFull(keys, types.KeyRefresh) {
		prepend = append(prepend, types.KeyRefresh)
	}

	if a.page.Get().GetSupportFiltering() && !types.KeyBindingContainsFull(keys, types.KeyFilter) {
		appended = append(appended, types.KeyFilter)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyCommander) {
		appended = append(appended, types.KeyCommander)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyHelp) {
		appended = append(appended, types.KeyHelp)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyQuit) {
		appended = append(appended, types.KeyQuit)
	}

	if len(keys) == 0 {
		keys = [][]key.Binding{{}}
	}
	keys[len(keys)-1] = append(keys[len(keys)-1], appended...)

	return append([][]key.Binding{prepend}, keys...)
}
