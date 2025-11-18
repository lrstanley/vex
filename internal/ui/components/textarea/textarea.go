// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package textarea

import (
	"charm.land/bubbles/v2/key"
	bta "charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/confirmable"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	_ types.Component                 = (*Model)(nil) // Ensure we implement the component interface.
	_ confirmable.Validatable[string] = (*Model)(nil) // Ensure we implement the validatable interface.
	_ confirmable.Focusable           = (*Model)(nil) // Ensure we implement the focusable interface.
)

// TODO:
//   - scrollbar
//   - separate textinput component, or make this support both?

// Model represents the textarea component.
type Model struct {
	types.ComponentModel

	// Core state.
	app            types.AppState
	needsScrollbar bool

	// Child components.
	textarea bta.Model
}

// New creates a new textarea component.
func New(app types.AppState, defaultValue string) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		textarea:       bta.New(),
	}

	m.textarea.Placeholder = "Enter text..."
	m.textarea.ShowLineNumbers = true
	m.textarea.Focus()
	m.textarea.Prompt = ""
	m.textarea.SetValue(defaultValue)

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	textareaStyles := bta.DefaultStyles(true)

	textareaStyles.Cursor.Blink = true
	textareaStyles.Cursor.Color = styles.Theme.AppCursor()

	textareaStyles.Focused.Text = lipgloss.NewStyle().
		Foreground(styles.Theme.AppFg()).
		Background(styles.Theme.AdaptAuto(styles.Theme.AppBg(), 0.05))
	textareaStyles.Blurred.Text = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), -0.15)).
		Background(styles.Theme.AdaptAuto(styles.Theme.AppBg(), 0.05))

	textareaStyles.Focused.CursorLine = lipgloss.NewStyle().
		Background(styles.Theme.AdaptAuto(styles.Theme.AppBg(), 0.15))
	textareaStyles.Blurred.CursorLine = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), -0.15)).
		Background(styles.Theme.AdaptAuto(styles.Theme.AppBg(), 0.05))

	textareaStyles.Focused.CursorLineNumber = lipgloss.NewStyle().
		Foreground(styles.Theme.AppFg())
	textareaStyles.Blurred.CursorLineNumber = lipgloss.NewStyle().
		Foreground(styles.Theme.AppFg())

	textareaStyles.Focused.LineNumber = lipgloss.NewStyle().
		Foreground(styles.Theme.AppBrightFg())
	textareaStyles.Blurred.LineNumber = lipgloss.NewStyle().
		Foreground(styles.Theme.AppBrightFg())

	textareaStyles.Focused.EndOfBuffer = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Blurred.EndOfBuffer = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))

	textareaStyles.Focused.Placeholder = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Blurred.Placeholder = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))

	textareaStyles.Focused.Prompt = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Blurred.Prompt = lipgloss.NewStyle().
		Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))

	m.textarea.SetStyles(textareaStyles)
}

func (m *Model) Init() tea.Cmd {
	return m.textarea.Focus()
}

// SetHeight sets the total height of the textarea.
func (m *Model) SetHeight(height int) {
	m.SetDimensions(m.Width, height)
}

// SetWidth sets the total width of the textarea.
func (m *Model) SetWidth(width int) {
	m.SetDimensions(width, m.Height)
}

// SetDimensions sets the dimensions of the component. Use this instead of
// [Model.SetWidth] or [Model.SetHeight] when possible to improve performance.
func (m *Model) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
	m.needsScrollbar = m.textarea.LineCount() > m.Height
	m.textarea.SetHeight(m.Height)
	if m.needsScrollbar {
		m.textarea.SetWidth(m.Width - 1) // -1=scrollbar.
	} else {
		m.textarea.SetWidth(m.Width)
	}
}

// Focus focuses the component.
func (m *Model) Focus() tea.Cmd {
	return m.textarea.Focus()
}

// Blur blurs the component.
func (m *Model) Blur() tea.Cmd {
	m.textarea.Blur()
	return nil
}

// Focused returns if the component has focus.
func (m *Model) Focused() bool {
	return m.textarea.Focused()
}

// HasInputFocus returns if the component has input focus.
func (m *Model) HasInputFocus() bool {
	return m.textarea.Focused()
}

// GetValue returns the current textarea value.
func (m *Model) GetValue() string {
	return m.textarea.Value()
}

// SetValue sets the textarea value.
func (m *Model) SetValue(value string) {
	m.textarea.SetValue(value)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetDimensions(msg.Width, msg.Height)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		return nil
	case tea.KeyPressMsg:
		switch {
		// TODO: textarea doesn't have page up/down at the moment.
		case key.Matches(msg, types.KeyPageUp):
			for range m.Height - 1 {
				m.textarea.CursorUp()
			}
			return nil
		case key.Matches(msg, types.KeyPageDown):
			for range m.Height - 1 {
				m.textarea.CursorDown()
			}
			return nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return cmd
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	if m.needsScrollbar {
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.textarea.View(),
			// TODO: this is wrong because it shouldn't be Line(), as thats the cursor, not the first visible line.
			styles.Scrollbar(
				m.Height,
				m.textarea.LineCount(),
				m.Height,
				m.textarea.ScrollYOffset(),
				styles.IconScrollbar,
				styles.IconScrollbar,
			),
		)
	}
	return m.textarea.View()
}
