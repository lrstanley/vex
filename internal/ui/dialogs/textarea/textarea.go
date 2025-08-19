// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package textarea

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/confirmable"
	"github.com/lrstanley/vex/internal/ui/components/textarea"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

// TODO:
//   - if dialog has input focus, don't render the keybinds in the dialog.

// Model represents the textarea dialog.
type Model struct {
	*types.DialogModel

	// Core state.
	app   types.AppState
	title string

	// Child components.
	textarea *confirmable.Model[*textarea.Model, string]

	// Styles.
	textareaStyle        lipgloss.Style
	focusedTextareaStyle lipgloss.Style
}

// New creates a new textarea dialog.
func New(app types.AppState, config confirmable.Config[string], title, defaultValue string) *Model {
	originalCancelFn := config.CancelFn
	config.CancelFn = func() tea.Cmd {
		if originalCancelFn != nil {
			return tea.Sequence(
				originalCancelFn(),
				types.CloseActiveDialog(),
			)
		}
		return types.CloseActiveDialog()
	}

	originalConfirmFn := config.ConfirmFn
	config.ConfirmFn = func(value string) tea.Cmd {
		if originalConfirmFn != nil {
			return tea.Sequence(
				originalConfirmFn(value),
				types.CloseActiveDialog(),
			)
		}
		return types.CloseActiveDialog()
	}

	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeLarge,
			DisableChildren: true,
		},
		app:      app,
		title:    title,
		textarea: confirmable.New(app, textarea.New(app, defaultValue), config),
	}

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.textareaStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Theme.DialogFg()).
		Margin(0, 1)

	m.focusedTextareaStyle = m.textareaStyle.
		BorderForeground(styles.Theme.DialogBorderFg())
}

func (m *Model) GetTitle() string {
	return m.title
}

func (m *Model) HasInputFocus() bool {
	return true
}

func (m *Model) Init() tea.Cmd {
	return m.textarea.Focus()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width
		m.textarea.SetDimensions(m.Width, m.Height)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyQuit):
			switch m.textarea.FocusedElement() { //nolint:exhaustive
			case confirmable.FocusWrapped:
				return m.textarea.SetFocus(confirmable.FocusCancel)
			default:
				return types.AppQuit()
			}
		case key.Matches(msg, types.KeyCancel):
			switch m.textarea.FocusedElement() { //nolint:exhaustive
			case confirmable.FocusCancel, confirmable.FocusNone:
				return types.CloseActiveDialog()
			default:
				return m.textarea.Update(msg)
			}
		}
	}

	return m.textarea.Update(msg)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.textarea.View()
}
