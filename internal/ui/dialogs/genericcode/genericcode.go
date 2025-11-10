// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package genericcode

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/viewport"
)

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel

	// Core state.
	app   types.AppState
	title string

	// Child components.
	code *viewport.Model
}

func New(app types.AppState, title, content, language string) *Model {
	m := &Model{
		DialogModel: &types.DialogModel{
			Size:          types.DialogSizeLarge,
			ShortKeyBinds: []key.Binding{types.KeyCopy},
			FullKeyBinds:  [][]key.Binding{{types.KeyCopy}},
		},
		app:   app,
		title: title,
	}

	m.code = viewport.New(app)
	m.code.SetCode(content, language)

	return m
}

func NewJSON(app types.AppState, title string, mask bool, data any) *Model {
	return New(app, title, formatter.ToJSON(data, mask, 2), "json")
}

func NewYAML(app types.AppState, title string, mask bool, data any) *Model {
	return New(app, title, formatter.ToYAML(data, mask, 2), "yaml")
}

func (m *Model) GetTitle() string {
	return m.title
}

func (m *Model) Init() tea.Cmd {
	return m.code.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width

		// If the viewport is smaller than the dialog height, resize the dialog
		// even smaller.
		m.code.SetDimensions(m.Width, m.Height)
		m.Height = min(m.code.TotalLineCount(), m.Height)
		m.code.SetHeight(m.Height)
		return nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyDetails) || key.Matches(msg, types.KeyToggleMask):
			return types.CloseActiveDialog()
		}
	}

	cmds = append(cmds, m.code.Update(msg))
	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.code.View()
}
