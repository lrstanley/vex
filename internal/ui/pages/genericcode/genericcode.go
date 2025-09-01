// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package genericcode

import (
	"encoding/json"
	"fmt"

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
			ShortKeyBinds:    []key.Binding{types.KeyCopy},
			FullKeyBinds:     [][]key.Binding{{types.KeyCopy}},
		},
		app:   app,
		title: title,
	}

	m.code = viewport.New(app)
	m.code.SetCode(content, language)

	return m
}

func NewJSON(app types.AppState, title string, data any) *Model {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return New(app, title, fmt.Sprintf("error: %v", err), "text")
	}
	return New(app, title, string(b), "json")
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
		m.code.SetDimensions(m.width, m.height)
		return nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyDetails):
			return types.CloseActivePage()
		case key.Matches(msg, types.KeyCancel):
			if m.app.Page().HasParent() {
				return types.CloseActivePage()
			}
			return nil
		case key.Matches(msg, types.KeyQuit):
			return types.AppQuit()
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
