// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package confirmable

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

var (
	_ Validatable[string] = (*mockWrapped)(nil) // Ensure we implement the validatable interface.
	_ Focusable           = (*mockWrapped)(nil) // Ensure we implement the focusable interface.
)

type mockWrapped struct {
	types.ComponentModel
}

func (m *mockWrapped) Focus() tea.Cmd {
	return nil
}

func (m *mockWrapped) Blur() tea.Cmd {
	return nil
}

func (m *mockWrapped) SetDimensions(_, _ int) {}

func (m *mockWrapped) GetValue() string {
	return ""
}

func (m *mockWrapped) HasInputFocus() bool {
	return false
}

func (m *mockWrapped) Init() tea.Cmd {
	return nil
}

func (m *mockWrapped) Update(_ tea.Msg) tea.Cmd {
	return nil
}

func (m *mockWrapped) View() string {
	return ""
}
