// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package statuselement

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/spinner"
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
	status     *types.StatusTextMsg
	operations []types.StatusOperationMsg

	// Styles.
	statusStyle    lipgloss.Style
	operationStyle lipgloss.Style

	// Child components.
	spinner spinner.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
		spinner:        spinner.New(),
	}

	m.spinner.Spinner = spinner.MiniDot

	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	m.spinner.Style = lipgloss.NewStyle().
		Foreground(styles.Theme.BarFg())

	m.statusStyle = lipgloss.NewStyle().
		Padding(0, 1)

	m.operationStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.InfoFg()).
		Background(styles.Theme.InfoBg()).
		Padding(0, 1)
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return nil
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	case spinner.TickMsg:
		if m.spinner.ID() != msg.ID {
			return nil
		}

		if m.status == nil && len(m.operations) == 0 {
			return nil
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case types.StatusMsg:
		switch msg := msg.Msg.(type) {
		case types.StatusTextMsg:
			if m.status != nil && (m.status.Status == types.Error || m.status.Status == types.Warning) && (msg.Status != types.Error && msg.Status != types.Warning) {
				// If the previous status was an error or warning, and the new status is
				// not an error or warning, skip the message that was just sent.
				return nil
			}
			m.status = &msg
			cmds = append(cmds, m.spinner.Tick)
		case types.ClearStatusTextMsg:
			if m.status != nil && (m.status.ID == msg.ID || msg.ID == 0) {
				m.status = nil
			}
		}
	case types.StatusOperationMsg:
		m.operations = append(m.operations, msg)
		cmds = append(cmds, m.spinner.Tick)
	case types.ClearStatusOperationMsg:
		for i, op := range m.operations {
			if op.ID == msg.ID {
				m.operations = append(m.operations[:i], m.operations[i+1:]...)
				break
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	switch {
	case len(m.operations) > 0:
		op := m.operations[len(m.operations)-1]
		return m.spinner.View() + m.operationStyle.
			Render(op.Text)
	case m.status != nil:
		fg, bg := styles.Theme.ByStatus(m.status.Status)

		return m.spinner.View() + m.statusStyle.
			Foreground(fg).
			Background(bg).
			Render(strings.TrimSpace(strings.Split(m.status.Text, "\n")[0]))
	default:
		return ""
	}
}
