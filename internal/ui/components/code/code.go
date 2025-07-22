// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package code

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	code     string
	language string
	content  string

	// Styles.
	baseStyle lipgloss.Style

	// Components.
	viewport viewport.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		viewport:       viewport.New(),
	}

	m.setStyles()
	m.viewport.FillHeight = true
	return m
}

func (m *Model) setStyles() {
	m.baseStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.Fg()).
		Background(styles.Theme.Bg())

	m.viewport.Style = lipgloss.NewStyle().
		Background(styles.Theme.Bg()).
		Padding(0, 1)
}

func (m *Model) SetCode(code, language string) {
	m.code = strings.TrimSpace(code)
	m.language = language
	m.renderCode()
}

func (m *Model) SetJSON(obj any) {
	out, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		m.SetError(fmt.Errorf("Error parsing JSON: %v", err))
		return
	}
	m.SetCode(string(out), "json")
}

func (m *Model) SetError(err error) {
	m.SetCode(fmt.Sprintf("Error: %v", err), "text")
}

func (m *Model) renderCode() {
	if m.code == "" {
		m.content = ""
		m.viewport.SetContent("")
		return
	}

	// Determine the lexer based on language.
	var lexer chroma.Lexer
	if m.language != "" {
		lexer = lexers.Get(m.language)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Use a simple style for syntax highligvexng.
	style := chromastyles.Fallback

	// Use a simple formatter that outputs ANSI color codes.
	formatter := formatters.TTY256

	// Tokenize the code.
	iterator, err := lexer.Tokenise(nil, m.code)
	if err != nil {
		m.content = m.code
		m.viewport.SetContent(m.code)
		return
	}

	// Format the tokens.
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		m.content = m.code
		m.viewport.SetContent(m.code)
		return
	}

	m.content = buf.String()
	m.viewport.SetContent(m.content)
}

func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
		m.viewport.SetHeight(m.Height)
		m.viewport.SetWidth(m.Width)
	case styles.ThemeUpdatedMsg:
		m.setStyles()
		m.renderCode()
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}

	return m.viewport.View()
}
