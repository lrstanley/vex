// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package alert

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
	"github.com/lrstanley/x/charm/formatter"
)

const minButtonWidth = 4

type Config struct {
	// Title is the title of the alert.
	Title string

	// Message is the main text body of the alert.
	Message string

	// ButtonText is the text of the button. Defaults to "ok".
	ButtonText string
}

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel

	// Core state.
	app    types.AppState
	config Config

	// Styles.
	messageStyle lipgloss.Style
	buttonStyle  lipgloss.Style
}

func New(app types.AppState, config Config) *Model {
	if config.ButtonText == "" {
		config.ButtonText = "ok"
	}

	config.ButtonText = formatter.PadMinimum(config.ButtonText, minButtonWidth)

	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeSmall,
			DisableChildren: true,
			ShortKeyBinds:   []key.Binding{types.OverrideHelp(types.KeySelectItem, "confirm")},
		},
		app:    app,
		config: config,
	}

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.messageStyle = lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(1)

	m.buttonStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.ActiveButtonFg()).
		Background(styles.Theme.ActiveButtonBg()).
		Padding(0, 2).
		Margin(0, 1)
}

func (m *Model) GetTitle() string {
	return m.config.Title
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width

		mh := strings.Count(
			ansi.Wrap(m.config.Message, m.Width-m.messageStyle.GetHorizontalFrameSize(), ""),
			"\n",
		)

		m.Height = mh + m.messageStyle.GetVerticalFrameSize() + 1 // +1 for the button.
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel, types.KeySelectItem, types.KeySelectItemAlt):
			return types.CloseActiveDialog()
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.messageStyle.Width(m.Width).Render(m.config.Message),
		lipgloss.PlaceHorizontal(
			m.Width,
			lipgloss.Right,
			m.buttonStyle.Render(m.config.ButtonText),
		),
	)
}
