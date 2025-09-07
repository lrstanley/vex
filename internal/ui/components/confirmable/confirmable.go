// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package confirmable

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const minButtonWidth = 6

// Focus represents which part of the component has focus.
type Focus int

const (
	FocusNone Focus = iota
	FocusWrapped
	FocusCancel
	FocusConfirm
)

// Focusable is a component that can be focused.
type Focusable interface {
	Focus() tea.Cmd
	Blur() tea.Cmd
	HasInputFocus() bool
}

// Validatable is a component that can be validated.
type Validatable[T any] interface {
	GetValue() T
}

// Wrapped is a component that can be wrapped.
type Wrapped interface {
	SetDimensions(width, height int)
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View() string
}

// Config holds the configuration for the textarea component.
type Config[T any] struct {
	// CancelText is the text of the cancel button. Defaults to "cancel".
	CancelText string

	// CancelFn if provided, will be called when the cancel button is pressed.
	CancelFn func() tea.Cmd

	// ConfirmText is the text of the confirm button. Defaults to "confirm".
	ConfirmText string

	// ConfirmStatus is the status of the confirm button. If not specified, will
	// default to active button colors based on the theme.
	ConfirmStatus types.Status

	// ConfirmFn if provided, will be called when the confirm button is pressed.
	// value is the value of the wrapped component, and is only provided if the
	// wrapped component is validatable (implements [Validatable]).
	ConfirmFn func(value T) tea.Cmd

	// PassthroughTab if true, will allow the tab key to be passed to the wrapped
	// component.
	PassthroughTab bool

	// Validator if provided, will be called when the confirm button is pressed,
	// and the wrapped component is validatable (implements [Validatable]). If
	// the validator returns an error, the confirm button will be be a no-op,
	// and the error will be displayed to the user.
	Validator func(value T) error
}

var _ types.Component = (*Model[*WrappedModel, string])(nil) // Ensure we implement the component interface.

type Model[T Wrapped, V any] struct {
	types.ComponentModel

	// Core state.
	app            types.AppState
	config         Config[V]
	focus          Focus
	validatorError error
	Wrapped        T

	// Styles.
	validatorErrorStyle lipgloss.Style
	buttonStyle         lipgloss.Style
	focusedButtonStyle  lipgloss.Style
}

func New[T Wrapped, V any](app types.AppState, wrapped T, config Config[V]) *Model[T, V] {
	if config.CancelText == "" {
		config.CancelText = "cancel"
	}
	if config.ConfirmText == "" {
		config.ConfirmText = "confirm"
	}

	config.CancelText = formatter.PadMinimum(config.CancelText, minButtonWidth)
	config.ConfirmText = formatter.PadMinimum(config.ConfirmText, minButtonWidth)

	m := &Model[T, V]{
		ComponentModel: types.ComponentModel{},
		app:            app,
		Wrapped:        wrapped,
		config:         config,
	}

	if _, canBeFocused := any(wrapped).(Focusable); canBeFocused {
		m.focus = FocusWrapped
	} else {
		m.focus = FocusCancel
	}

	m.initStyles()
	return m
}

func (m *Model[T, V]) initStyles() {
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
}

func (m *Model[T, V]) validate() error {
	wv, canBeValidated := any(m.Wrapped).(Validatable[V])
	if m.config.Validator != nil && canBeValidated {
		m.validatorError = m.config.Validator(wv.GetValue())
	} else {
		m.validatorError = nil
	}
	return m.validatorError
}

// SetHeight sets the total height of the component.
func (m *Model[T, V]) SetHeight(height int) {
	m.SetDimensions(m.Width, height)
}

// SetWidth sets the total width of the component.
func (m *Model[T, V]) SetWidth(width int) {
	m.SetDimensions(width, m.Height)
}

// SetDimensions sets the dimensions of the component. Use this instead of
// [Model.SetWidth] or [Model.SetHeight] when possible to improve performance.
func (m *Model[T, V]) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
	m.Wrapped.SetDimensions(m.Width, m.Height-1) // -1=buttons at bottom.
}

// Focused returns true if the component has focus.
func (m *Model[T, V]) Focused() bool {
	return m.focus != FocusNone
}

func (m *Model[T, V]) FocusedElement() Focus {
	return m.focus
}

// Focus sets the focus to the wrapped component.
func (m *Model[T, V]) Focus() tea.Cmd {
	return m.SetFocus(FocusWrapped)
}

// Blur sets the focus to none.
func (m *Model[T, V]) Blur() tea.Cmd {
	return m.SetFocus(FocusNone)
}

// SetFocus sets the focus to the given focus.
func (m *Model[T, V]) SetFocus(focus Focus) tea.Cmd {
	if m.focus == focus {
		return nil
	}

	wf, canBeFocused := any(m.Wrapped).(Focusable)

	if !canBeFocused && focus == FocusWrapped {
		focus = FocusCancel
	}

	m.focus = focus
	switch focus {
	case FocusNone, FocusCancel, FocusConfirm:
		if canBeFocused {
			return wf.Blur()
		}
		return nil
	case FocusWrapped:
		if canBeFocused {
			return wf.Focus()
		}
		return nil
	}
	return nil
}

// TabForward moves the focus to the next component in the order of FocusWrapped,
// FocusCancel, FocusConfirm.
func (m *Model[T, V]) TabForward() tea.Cmd {
	switch m.focus {
	case FocusNone:
		return m.SetFocus(FocusWrapped)
	case FocusWrapped:
		return m.SetFocus(FocusCancel)
	case FocusCancel:
		return m.SetFocus(FocusConfirm)
	case FocusConfirm:
		return m.SetFocus(FocusWrapped)
	}
	return nil
}

// TabBackward moves the focus to the previous component in the order of FocusWrapped,
// FocusCancel, FocusConfirm.
func (m *Model[T, V]) TabBackward() tea.Cmd {
	switch m.focus {
	case FocusNone:
		return m.SetFocus(FocusWrapped)
	case FocusWrapped:
		return m.SetFocus(FocusConfirm)
	case FocusCancel:
		return m.SetFocus(FocusWrapped)
	case FocusConfirm:
		return m.SetFocus(FocusCancel)
	}
	return nil
}

func (m *Model[T, V]) Init() tea.Cmd {
	return m.Wrapped.Init()
}

func (m *Model[T, V]) Update(msg tea.Msg) tea.Cmd {
	var wrappedFocusedWithInput bool
	if wf, canBeFocused := any(m.Wrapped).(Focusable); canBeFocused {
		wrappedFocusedWithInput = m.focus == FocusWrapped && wf.HasInputFocus()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetDimensions(msg.Width, msg.Height)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		return m.Wrapped.Update(msg)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel):
			switch m.focus { //nolint:exhaustive
			case FocusWrapped:
				return m.SetFocus(FocusCancel)
			default:
				if m.config.CancelFn != nil {
					return m.config.CancelFn()
				}
				return nil
			}
		case key.Matches(msg, types.KeyTabForward):
			if m.focus == FocusWrapped && m.config.PassthroughTab {
				return m.Wrapped.Update(msg)
			}
			return m.TabForward()
		case key.Matches(msg, types.KeyTabBackward):
			if m.focus == FocusWrapped && m.config.PassthroughTab {
				return m.Wrapped.Update(msg)
			}
			return m.TabBackward()
		case !wrappedFocusedWithInput && key.Matches(msg, types.KeySelectItem):
			switch m.focus { //nolint:exhaustive
			case FocusCancel:
				if m.config.CancelFn != nil {
					return m.config.CancelFn()
				}
			case FocusConfirm:
				if err := m.validate(); err != nil {
					return types.SendStatus("invalid input: "+err.Error(), types.Error, 2*time.Second)
				}

				if m.config.ConfirmFn != nil {
					wv, canBeValidated := any(m.Wrapped).(Validatable[V])
					if canBeValidated {
						return m.config.ConfirmFn(wv.GetValue())
					}
					var v V
					return m.config.ConfirmFn(v)
				}
			}
			return nil
		case !wrappedFocusedWithInput && (key.Matches(msg, types.KeyLeft) || key.Matches(msg, types.KeyUp)):
			return m.TabBackward()
		case !wrappedFocusedWithInput && (key.Matches(msg, types.KeyRight) || key.Matches(msg, types.KeyDown)):
			return m.TabForward()
		case wrappedFocusedWithInput:
			m.validatorError = nil
			return m.Wrapped.Update(msg)
		}
	default:
		return m.Wrapped.Update(msg)
	}
	return nil
}

func (m *Model[T, V]) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	var cancel, confirm string

	switch m.focus { //nolint:exhaustive
	case FocusCancel:
		cancel = m.focusedButtonStyle.Render(m.config.CancelText)
		confirm = m.buttonStyle.Render(m.config.ConfirmText)
	case FocusConfirm:
		cancel = m.buttonStyle.Render(m.config.CancelText)
		if m.config.ConfirmStatus == "" {
			confirm = m.focusedButtonStyle.Render(m.config.ConfirmText)
		} else {
			fg, bg := styles.Theme.ByStatus(m.config.ConfirmStatus)
			confirm = m.focusedButtonStyle.
				Foreground(fg).
				Background(bg).
				Render(m.config.ConfirmText)
		}
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
		m.Wrapped.View(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			validatorError,
			cancel,
			confirm,
		),
	)
}
