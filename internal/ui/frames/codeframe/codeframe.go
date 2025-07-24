// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package codeframe

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/code"
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height   int
	width    int
	title    string
	content  string
	language string

	// Child components.
	codeComponent *code.Model
}

func New(app types.AppState, title, content, language string) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         []string{},
			SupportFiltering: false,
			ShortKeyBinds:    []key.Binding{types.KeyCancel, types.KeyQuit},
			FullKeyBinds:     [][]key.Binding{{types.KeyCancel, types.KeyQuit}},
		},
		app:      app,
		title:    title,
		content:  content,
		language: language,
	}

	m.codeComponent = code.New(app)

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.codeComponent.Init(),
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

	// Set the code content after initialization.
	if m.content != "" {
		m.codeComponent.SetCode(m.content, m.language)
		m.content = "" // Clear to avoid setting repeatedly.
	}

	cmds = append(cmds, m.codeComponent.Update(msg))
	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return m.codeComponent.View()
}

func (m *Model) GetTitle() string {
	return m.title
}
