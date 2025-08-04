// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package state

import (
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
