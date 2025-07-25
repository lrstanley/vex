// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package codeframe

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/viewport"
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height int
	width  int
	title  string

	// Child components.
	code *viewport.Model
}

func New(app types.AppState, title, content, language string) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         []string{},
			SupportFiltering: false,
			ShortKeyBinds:    []key.Binding{types.KeyCancel, types.KeyQuit},
			FullKeyBinds:     [][]key.Binding{{types.KeyCancel, types.KeyQuit}},
		},
		app:   app,
		title: title,
	}

	m.code = viewport.New(app)
	m.code.SetCode(content, language)

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.code.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel):
			if m.app.Page().HasParent() {
				return types.CloseActivePage()
			}
			return nil
		case key.Matches(msg, types.KeyQuit):
			return tea.Quit
		}
	}

	cmds = append(cmds, m.code.Update(msg))
	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return m.code.View()
}

func (m *Model) GetTitle() string {
	return m.title
}
