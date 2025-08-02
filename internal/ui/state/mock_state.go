// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package state

import "github.com/lrstanley/vex/internal/types"

type mockPage struct {
	types.PageModel
}

func (m *mockPage) GetTitle() string {
	return "mock-page"
}

func NewMockAppState(client types.Client, initialPage types.Page) *AppState {
	if initialPage == nil {
		initialPage = &mockPage{} // Stub.
	}
	return &AppState{
		page:   NewPageState(initialPage),
		dialog: NewDialogState(),
		client: client,
	}
}
