// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package shorthelp

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type Styles struct {
	Base           lipgloss.Style
	Desc           lipgloss.Style
	Key            lipgloss.Style
	Separator      lipgloss.Style
	EllipsisChar   string
	SeparatorChars string
}

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	help string
	kb   []key.Binding

	// Styles.
	Styles Styles
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
	}

	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	if m.Styles.EllipsisChar == "" {
		m.Styles.EllipsisChar = styles.IconEllipsis
	}

	if m.Styles.SeparatorChars == "" {
		m.Styles.SeparatorChars = " " + styles.IconSeparator + " "
	}

	m.Styles.Desc = m.Styles.Desc.Inherit(m.Styles.Base)
	m.Styles.Key = m.Styles.Key.Inherit(m.Styles.Base)
	m.Styles.Separator = m.Styles.Separator.Inherit(m.Styles.Base)
}

func (m *Model) SetKeyBinds(kb ...key.Binding) {
	m.kb = kb
	m.generateShortHelp()
}

func (m *Model) generateShortHelp() {
	var s strings.Builder

	for i, kb := range m.kb {
		if !kb.Enabled() {
			continue
		}

		if i > 0 {
			s.WriteString(m.Styles.Separator.Render(m.Styles.SeparatorChars))
		}

		s.WriteString(
			m.Styles.Key.Render(kb.Help().Key) +
				m.Styles.Key.Render(" ") +
				m.Styles.Desc.Render(kb.Help().Desc),
		)
	}

	m.help = s.String()
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		m.setStyles()
		return nil
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg.(type) {
	case styles.ThemeUpdatedMsg:
		m.setStyles()
		m.generateShortHelp()
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	// TODO: bundle truncation logic.
	return m.help
}
