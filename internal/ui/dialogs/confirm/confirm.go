// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package confirm

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/confirmable"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type confirmableModel struct {
	confirmable.WrappedModel
	model *Model
}

func (m *confirmableModel) View() string {
	return m.model.messageStyle.Width(m.model.Width).Render(m.model.config.Message)
}

type Config struct {
	// Title is the title of the dialog.
	Title string

	// Message is the main text body of the dialog.
	Message string

	// AllowsBlur is whether the dialog allows [types.KeyCancel] to be used to
	// close the dialog.
	AllowsBlur bool

	// CancelText is the text of the cancel button. Defaults to "cancel".
	CancelText string

	// CancelFn is the function to call when the cancel button is pressed, or
	// [types.KeyCancel] is pressed.
	CancelFn func() tea.Cmd

	// ConfirmText is the text of the confirm button. Defaults to "confirm".
	ConfirmText string

	// ConfirmStatus is the status of the confirm button. If not specified, will
	// default to active button colors based on the theme.
	ConfirmStatus types.Status

	// ConfirmFn is the function to call when the confirm button is pressed.
	ConfirmFn func() tea.Cmd
}

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel

	// Core state.
	app    types.AppState
	config Config

	// Styles.
	messageStyle lipgloss.Style

	// Child components.
	confirmable *confirmable.Model[*confirmableModel, struct{}]
}

func New(app types.AppState, config Config) *Model {
	cc := confirmable.Config[struct{}]{
		CancelText:    config.CancelText,
		CancelFn:      config.CancelFn,
		ConfirmText:   config.ConfirmText,
		ConfirmStatus: config.ConfirmStatus,
		ConfirmFn: func(_ struct{}) tea.Cmd {
			if config.ConfirmFn != nil {
				return config.ConfirmFn()
			}
			return nil
		},
	}

	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeSmall,
			DisableChildren: true,
			ShortKeyBinds:   []key.Binding{types.OverrideHelp(types.KeySelectItem, "confirm")},
		},
		app:    app,
		config: config,
	}

	actualModel := &confirmableModel{model: m}
	m.confirmable = confirmable.New(app, actualModel, cc)

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.messageStyle = lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(1)
}

func (m *Model) GetTitle() string {
	return m.config.Title
}

func (m *Model) HasInputFocus() bool {
	return true
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width

		mh := strings.Count(
			ansi.Wrap(m.config.Message, m.Width-m.messageStyle.GetHorizontalFrameSize(), ""),
			"\n",
		)

		m.Height = mh + m.messageStyle.GetVerticalFrameSize() + 1 // +1 for the button.
		m.confirmable.SetDimensions(m.Width, m.Height)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyQuit):
			return types.AppQuit()
		case key.Matches(msg, types.KeyCancel) && !m.config.AllowsBlur:
			return nil
		}
	}

	return m.confirmable.Update(msg)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.confirmable.View()
}
