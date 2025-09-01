// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package confirmable

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

var _ Wrapped = (*WrappedModel)(nil) // Ensure we implement the wrapped interface.

// WrappedModel is a basic model you can use as a base for your model, to implement
// some of the basic [Wrapped] interface.
type WrappedModel struct{}

func (m *WrappedModel) SetDimensions(_, _ int) {}

func (m *WrappedModel) Init() tea.Cmd {
	return nil
}

func (m *WrappedModel) Update(_ tea.Msg) tea.Cmd {
	return nil
}

func (m *WrappedModel) View() string {
	return ""
}
