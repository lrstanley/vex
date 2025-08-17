// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package textarea

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textarea"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

// TODO:
//   - if dialog has input focus, don't render the keybinds in the dialog.
//   - scrollbar
//   - separate textinput component, or make this support both?

// Focus represents which part of the dialog has focus.
type Focus int

const (
	FocusTextarea Focus = iota
	FocusCancel
	FocusConfirm
)

// Config holds the configuration for the textarea dialog.
type Config struct {
	Title string

	// CancelText is the text of the cancel button. Defaults to "cancel".
	CancelText string

	// CancelFn if provided, will be called when the cancel button is pressed.
	CancelFn func() tea.Cmd

	// ConfirmText is the text of the confirm button. Defaults to "confirm".
	ConfirmText string

	// ConfirmFn if provided, will be called when the confirm button is pressed.
	ConfirmFn func(value string) tea.Cmd

	// Validator if provided, will be called when the confirm button is pressed.
	// If the validator returns an error, the confirm button will be be a no-op,
	// and the error will be displayed to the user.
	Validator func(value string) error

	// DefaultValue is the default value of the textarea.
	DefaultValue string
}

// Model represents the textarea dialog.
type Model struct {
	*types.DialogModel

	// Core state.
	app            types.AppState
	config         Config
	focus          Focus
	validatorError error

	// Child components.
	textarea textarea.Model

	// Styles.
	textareaStyle        lipgloss.Style
	focusedTextareaStyle lipgloss.Style
	validatorErrorStyle  lipgloss.Style
	buttonStyle          lipgloss.Style
	focusedButtonStyle   lipgloss.Style
}

// New creates a new textarea dialog.
func New(app types.AppState, config Config) *Model {
	if config.CancelText == "" {
		config.CancelText = "cancel"
	}
	if config.ConfirmText == "" {
		config.ConfirmText = "confirm"
	}

	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeLarge,
			DisableChildren: true,
		},
		app:      app,
		config:   config,
		focus:    FocusTextarea,
		textarea: textarea.New(),
	}

	m.textarea.Placeholder = "Enter text..."
	m.textarea.ShowLineNumbers = true
	m.textarea.Focus()
	m.textarea.Prompt = ""
	m.textarea.SetValue(config.DefaultValue)

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

	m.validatorErrorStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.ErrorFg()).
		Background(styles.Theme.ErrorBg()).
		Padding(0, 1).
		Margin(0, 1)

	m.buttonStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.InactiveButtonFg()).
		Background(styles.Theme.InactiveButtonBg()).
		Padding(0, 2).
		MarginRight(1)

	m.focusedButtonStyle = m.buttonStyle.
		Foreground(styles.Theme.ActiveButtonFg()).
		Background(styles.Theme.ActiveButtonBg())

	textareaStyles := textarea.DefaultStyles(true) // TODO
	textareaStyles.Cursor.Blink = true
	textareaStyles.Cursor.Color = styles.Theme.AppCursor()
	textareaStyles.Focused.Text = lipgloss.NewStyle().Foreground(styles.Theme.AppFg())
	textareaStyles.Blurred.Text = lipgloss.NewStyle().Foreground(styles.Theme.AppFg())
	textareaStyles.Focused.CursorLine = lipgloss.NewStyle().Background(styles.Theme.AdaptAuto(styles.Theme.AppBg(), 0.15))
	textareaStyles.Blurred.CursorLine = lipgloss.NewStyle()
	textareaStyles.Focused.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.Theme.AppFg())
	textareaStyles.Blurred.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.Theme.AppFg())
	textareaStyles.Focused.LineNumber = lipgloss.NewStyle().Foreground(styles.Theme.AppBrightFg())
	textareaStyles.Blurred.LineNumber = lipgloss.NewStyle().Foreground(styles.Theme.AppBrightFg())
	textareaStyles.Focused.EndOfBuffer = lipgloss.NewStyle().Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Blurred.EndOfBuffer = lipgloss.NewStyle().Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Focused.Placeholder = lipgloss.NewStyle().Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Focused.Prompt = lipgloss.NewStyle().Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	textareaStyles.Blurred.Prompt = lipgloss.NewStyle().Foreground(styles.Theme.AdaptAuto(styles.Theme.AppFg(), 0.2))
	m.textarea.SetStyles(textareaStyles)
}

func (m *Model) GetTitle() string {
	return m.config.Title
}

func (m *Model) HasInputFocus() bool {
	return true
}

func (m *Model) validate() error {
	if m.config.Validator != nil {
		m.validatorError = m.config.Validator(m.textarea.Value())
	} else {
		m.validatorError = nil
	}
	return m.validatorError
}

func (m *Model) Init() tea.Cmd {
	return m.textarea.Focus()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width
		// Adjust textarea size to account for sidebar
		m.textarea.SetWidth(m.Width - m.textareaStyle.GetHorizontalFrameSize())
		m.textarea.SetHeight(m.Height - m.textareaStyle.GetVerticalFrameSize() - 1) // -1 for buttons at bottom.
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		return nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel) || key.Matches(msg, types.KeyQuit):
			if m.focus == FocusTextarea {
				// If textarea has focus, move focus to cancel button.
				m.focus = FocusCancel
				m.textarea.Blur()
				return nil
			}

			if key.Matches(msg, types.KeyQuit) {
				return types.AppQuit()
			}

			// If focus already not in textarea, execute cancel function.
			if m.config.CancelFn != nil {
				return tea.Batch(
					m.config.CancelFn(),
					types.CloseActiveDialog(),
				)
			}
			return types.CloseActiveDialog()
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))): // TODO: proper keybinding.
			switch m.focus {
			case FocusTextarea:
				m.focus = FocusCancel
				m.textarea.Blur()
			case FocusCancel:
				m.focus = FocusConfirm
			case FocusConfirm:
				m.focus = FocusTextarea
				return m.textarea.Focus()
			}
			return nil
		case m.focus != FocusTextarea && key.Matches(msg, types.KeyLeft):
			if m.focus == FocusConfirm {
				m.focus = FocusCancel
			}
			return nil
		case m.focus != FocusTextarea && key.Matches(msg, types.KeyRight):
			if m.focus == FocusCancel {
				m.focus = FocusConfirm
			}
			return nil
		case key.Matches(msg, types.KeySelectItem):
			switch m.focus { //nolint:exhaustive
			case FocusCancel:
				if m.config.CancelFn != nil {
					return tea.Batch(
						m.config.CancelFn(),
						types.CloseActiveDialog(),
					)
				}
				return types.CloseActiveDialog()
			case FocusConfirm:
				if err := m.validate(); err != nil {
					return types.SendStatus("invalid input: "+err.Error(), types.Error, 2*time.Second)
				}

				if m.config.ConfirmFn != nil {
					return tea.Batch(
						m.config.ConfirmFn(m.textarea.Value()),
						types.CloseActiveDialog(),
					)
				}
				return types.CloseActiveDialog()
			}
		}
	}

	// Update textarea if it has focus.
	if m.focus == FocusTextarea {
		previousValue := m.textarea.Value()
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

		if previousValue != m.textarea.Value() {
			m.validatorError = nil
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	var textareaView string
	if m.focus == FocusTextarea {
		textareaView = m.focusedTextareaStyle.Render(m.textarea.View())
	} else {
		textareaView = m.textareaStyle.Render(m.textarea.View())
	}

	var cancel, confirm string

	switch m.focus { //nolint:exhaustive
	case FocusCancel:
		cancel = m.focusedButtonStyle.Render(m.config.CancelText)
		confirm = m.buttonStyle.Render(m.config.ConfirmText)
	case FocusConfirm:
		cancel = m.buttonStyle.Render(m.config.CancelText)
		confirm = m.focusedButtonStyle.Render(m.config.ConfirmText)
	default:
		cancel = m.buttonStyle.Render(m.config.CancelText)
		confirm = m.buttonStyle.Render(m.config.ConfirmText)
	}

	available := m.Width - ansi.StringWidth(cancel) - ansi.StringWidth(confirm)

	var validatorError string
	if m.validatorError != nil {
		validatorError = m.validatorErrorStyle.
			Render(formatter.Trunc(
				"invalid input: "+m.validatorError.Error(),
				max(0, available-m.validatorErrorStyle.GetHorizontalFrameSize()),
			))
		available -= ansi.StringWidth(validatorError)
		if available > 0 {
			validatorError += strings.Repeat(" ", available)
		}
	} else {
		validatorError = strings.Repeat(" ", max(0, available))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		textareaView,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			validatorError,
			cancel,
			confirm,
		),
	)
}

// GetValue returns the current textarea value.
func (m *Model) GetValue() string {
	return m.textarea.Value()
}

// SetValue sets the textarea value.
func (m *Model) SetValue(value string) {
	m.textarea.SetValue(value)
}
