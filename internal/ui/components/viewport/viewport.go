// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package viewport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/bubbles/v2/key"
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
	code         string // Code content. Stored so we can re-render styled code on theme change.
	language     string // Code language. Stored so we can re-render styled code on theme change.
	content      string // Non-code focused content. Stored so we can copy it.
	hasScrollbar bool   // Whether to show a scrollbar.

	// Components.
	viewport viewport.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		viewport:       viewport.New(),
	}

	m.viewport.FillHeight = true
	m.viewport.SoftWrap = true
	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	m.viewport.Style = lipgloss.NewStyle().
		Padding(0, 1)
}

func (m *Model) GotoTop() {
	m.viewport.GotoTop()
}

func (m *Model) GotoBottom() {
	m.viewport.GotoBottom()
}

func (m *Model) TotalLineCount() int {
	return m.viewport.TotalLineCount()
}

func (m *Model) VisibleLineCount() int {
	return m.viewport.VisibleLineCount()
}

func (m *Model) GetContent() string {
	if m.code != "" {
		return m.code
	}
	return m.content
}

func (m *Model) SetContent(content string) {
	m.code = ""
	m.language = ""
	m.viewport.SetContent(content)
	m.ensureSize()
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
		m.viewport.SetContent("")
		m.ensureSize()
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

	style := chroma.MustNewStyle("vex", styles.Theme.Chroma())

	// Use a simple formatter that outputs ANSI color codes.
	formatter := formatters.TTY256

	// Tokenize the code.
	iterator, err := lexer.Tokenise(nil, m.code)
	if err != nil {
		m.viewport.SetContent(m.code)
		m.ensureSize()
		return
	}

	// Format the tokens.
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		m.viewport.SetContent(m.code)
		m.ensureSize()
		return
	}

	m.viewport.SetContent(buf.String())
	m.ensureSize()
}

func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetHeight(msg.Height)
		m.SetWidth(msg.Width)
		return nil
	case styles.ThemeUpdatedMsg:
		m.setStyles()
		if m.code != "" {
			m.renderCode()
		}
		return nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCopy):
			return types.SetClipboard(m.GetContent())
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model) SetHeight(h int) {
	m.Height = h
	m.ensureSize()
	m.viewport.GotoTop()
}

func (m *Model) SetWidth(w int) {
	m.Width = w
	m.ensureSize()
}

func (m *Model) ensureSize() {
	m.viewport.SetHeight(m.Height)

	if m.viewport.TotalLineCount() > m.Height {
		m.hasScrollbar = true
		m.viewport.SetWidth(m.Width - styles.ScrollbarWidth)
	} else {
		m.hasScrollbar = false
		m.viewport.SetWidth(m.Width)
	}
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}

	if m.hasScrollbar {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.viewport.View(),
			styles.Scrollbar(
				m.Height,
				m.viewport.TotalLineCount(),
				m.viewport.VisibleLineCount(),
				m.viewport.YOffset,
			),
		)
	}

	return m.viewport.View()
}
