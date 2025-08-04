// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package state

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/lrstanley/vex/internal/types"
)

var _ types.AppState = &AppState{} // Ensure that appState implements the AppState interface.

type AppState struct {
	page   types.PageState
	dialog types.DialogState
	client types.Client
}

func (a *AppState) SetPage(page types.PageState) {
	a.page = page
}

func (a *AppState) Page() types.PageState {
	return a.page
}

func (a *AppState) SetDialog(dialog types.DialogState) {
	a.dialog = dialog
}

func (a *AppState) Dialog() types.DialogState {
	return a.dialog
}

func (a *AppState) SetClient(client types.Client) {
	a.client = client
}

func (a *AppState) Client() types.Client {
	return a.client
}

func (a *AppState) ShortHelp(focused types.FocusID) []key.Binding {
	keys := a.page.Get().ShortHelp()

	var prepended []key.Binding

	switch focused {
	case types.FocusDialog:
		dialog := a.dialog.Get(true)
		if dialog != nil {
			keys = append(dialog.ShortHelp(), keys...)
		}
	case types.FocusPage:
		if a.page.Get().GetSupportFiltering() && !types.KeyBindingContains(keys, types.KeyFilter) {
			prepended = append(prepended, types.KeyFilter)
		}

		if !types.KeyBindingContains(keys, types.KeyCommander) {
			prepended = append(prepended, types.KeyCommander)
		}
	}

	if !types.KeyBindingContains(keys, types.KeyHelp) {
		prepended = append(prepended, types.KeyHelp)
	}

	if !types.KeyBindingContains(keys, types.KeyQuit) {
		keys = append(keys, types.KeyQuit) // Add to the end.
	}

	return append(prepended, keys...)
}

func (a *AppState) FullHelp(focused types.FocusID) [][]key.Binding {
	keys := a.page.Get().FullHelp()

	var prepended, appended []key.Binding

	switch focused {
	case types.FocusDialog:
		dialog := a.dialog.Get(true)
		if dialog != nil {
			keys = append(dialog.FullHelp(), keys...)
		}
		prepended = append(prepended, types.KeyCancel)
	case types.FocusPage:
		if a.page.HasParent() && !types.KeyBindingContainsFull(keys, types.KeyCancel) {
			prepended = append(prepended, types.KeyCancel)
		}

		if a.page.Get().GetRefreshInterval() > 0 && !types.KeyBindingContainsFull(keys, types.KeyRefresh) {
			prepended = append(prepended, types.KeyRefresh)
		}

		if a.page.Get().GetSupportFiltering() && !types.KeyBindingContainsFull(keys, types.KeyFilter) {
			appended = append(appended, types.KeyFilter)
		}

		if !types.KeyBindingContainsFull(keys, types.KeyCommander) {
			appended = append(appended, types.KeyCommander)
		}
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

	keys[len(keys)-1] = append(keys[len(keys)-1], prepended...)

	return append(keys, appended)
}
