// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package errorview

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const (
	MaxWidth         = 60
	NotShownTemplate = styles.IconDanger + " %d error(s) not shown (screen size)"
)

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// UI state.
	errors   []string
	maxWidth int

	// Styles.
	titleStyle lipgloss.Style
	errorStyle lipgloss.Style
}

func New() *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
	}

	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	m.titleStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.ErrorFg()).
		Padding(0, 1)
	m.errorStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.ErrorFg()).
		Background(styles.Theme.ErrorBg()).
		Padding(0, 1)
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg.(type) {
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	}

	return tea.Batch(cmds...)
}

func (m *Model) SetHeight(height int) {
	m.Height = height
}

func (m *Model) SetWidth(width int) {
	m.Width = width
	m.calculateMaxWidth()
}

func (m *Model) SetErrors(errs ...error) {
	m.errors = make([]string, 0, len(errs))
	for _, err := range errs {
		e := strings.TrimSpace(err.Error())
		if slices.Contains(m.errors, e) {
			continue
		}
		m.errors = append(m.errors, e)
	}
	m.calculateMaxWidth()
}

func (m *Model) calculateMaxWidth() {
	if m.Width < MaxWidth {
		m.maxWidth = m.Width
		return
	}
	var w int
	for _, err := range m.errors {
		ww := lipgloss.Width(err) + m.errorStyle.GetHorizontalFrameSize() + 2 // +2=icon+space.
		if w < ww {
			w = ww
		}
	}
	m.maxWidth = min(MaxWidth, max(len(NotShownTemplate), w))
}

func (m *Model) View() string {
	if m.Height == 0 || m.Width == 0 {
		return ""
	}

	out := make([]string, 0, len(m.errors)+1)

	out = append(out, m.titleStyle.Render(styles.Title(
		fmt.Sprintf("%d errors", len(m.errors)),
		m.maxWidth-m.titleStyle.GetHorizontalFrameSize(),
		styles.IconTitleGradientDivider,
		styles.Theme.ErrorFg(),
		styles.Theme.ErrorFg(),
		styles.Theme.WarningFg(),
	)))

	h := styles.H(out...)

	for i, err := range m.errors {
		e := m.errorStyle.
			Width(m.maxWidth).
			Render(styles.IconDanger + " " + err)

		if eh := lipgloss.Height(e); eh+h > m.Height {
			out = append(
				out,
				m.errorStyle.
					Width(m.maxWidth).
					Render(fmt.Sprintf(NotShownTemplate, len(m.errors)-i)),
			)
			break
		} else {
			h += eh
		}

		out = append(out, e)
		if i < len(m.errors)-1 {
			out = append(out, "")
		}
	}

	return lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(lipgloss.JoinVertical(lipgloss.Center, out...))
}

// unfoldErrors unwraps errors using [errors.Unwrap].
func unfoldErrors(errs ...error) (out []error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		out = append(out, err)
		out = append(out, unfoldErrors(errors.Unwrap(err))...)
	}
	return out
}
