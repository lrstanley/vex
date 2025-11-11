// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package shorthelp

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
	"github.com/lrstanley/x/charm/formatter"
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

	// UI state.
	help string
	kb   []key.Binding

	// Styles.
	styles Styles
}

func New() *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
	}

	m.setStyles()
	return m
}

func NewWithKeyBinds(kb ...key.Binding) *Model {
	m := New()
	m.SetKeyBinds(kb...)
	return m
}

func (m *Model) setStyles() {
	if m.styles.EllipsisChar == "" {
		m.styles.EllipsisChar = styles.IconEllipsis
	}

	if m.styles.SeparatorChars == "" {
		m.styles.SeparatorChars = " " + styles.IconSeparator + " "
	}

	m.styles.Key = m.styles.Key.Inherit(m.styles.Base)
	m.styles.Desc = m.styles.Desc.Inherit(m.styles.Base)
	m.styles.Separator = m.styles.Separator.Inherit(m.styles.Base)
}

func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.setStyles()
}

func (m *Model) SetKeyBinds(kb ...key.Binding) {
	m.kb = kb
	m.generateShortHelp()
}

func (m *Model) NumKeyBinds() int {
	return len(m.kb)
}

func (m *Model) SetMaxWidth(width int) {
	cur := m.Width
	m.Width = width
	if cur != width {
		m.generateShortHelp()
	}
}

func (m *Model) generateShortHelp() {
	var s strings.Builder

	for i, kb := range m.kb {
		if !kb.Enabled() {
			continue
		}

		if i > 0 {
			s.WriteString(m.styles.Separator.Render(m.styles.SeparatorChars))
		}

		s.WriteString(
			m.styles.Key.Render(kb.Help().Key) +
				m.styles.Key.Render(" ") +
				m.styles.Desc.Render(kb.Help().Desc),
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
	case tea.WindowSizeMsg:
		return nil
	case styles.ThemeUpdatedMsg:
		m.setStyles()
		m.generateShortHelp()
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.Width > 0 {
		return m.styles.Base.Render(formatter.Trunc(
			m.help,
			m.Width-m.styles.Base.GetHorizontalFrameSize(),
		))
	}
	return m.styles.Base.Render(m.help)
}
