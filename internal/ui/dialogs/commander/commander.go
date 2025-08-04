// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package commander

import (
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/dialogselector"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type PageRef struct {
	Description string
	Commands    []string
	New         func() types.Page
}

type Config struct {
	App   types.AppState
	Pages []PageRef
}

func (c Config) Suggestions() (cmds []string) {
	for _, ref := range c.Pages {
		cmds = append(cmds, ref.Commands...)
	}
	return cmds
}

func (c Config) Len() int {
	return len(c.Pages)
}

func (c Config) GetData() (cols []string, rows [][]string) {
	cols = []string{"  Command", "Aliases", "Description"}

	currentPageCmds := c.App.Page().Get().GetCommands()

	for _, ref := range c.Pages {
		var name string

		var isCurrent bool
		for _, cmd := range ref.Commands {
			if slices.Contains(currentPageCmds, cmd) {
				isCurrent = true
				break
			}
		}
		if isCurrent {
			name = styles.IconClosedCircle + " " + ref.Commands[0]
		} else {
			name = "  " + ref.Commands[0]
		}
		rows = append(rows, []string{
			name,
			strings.Join(ref.Commands[1:], ", "),
			ref.Description,
		})
	}

	return cols, rows
}

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel

	// Core state.
	app types.AppState

	// UI state.
	config Config

	// Child components.
	selector *dialogselector.Model
}

func New(app types.AppState, config Config) *Model {
	for _, ref := range config.Pages {
		if ref.New == nil {
			panic("New function is required for all pages")
		}
		if len(ref.Commands) == 0 {
			panic("Commands are required for all pages")
		}
	}

	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeSmall,
			DisableChildren: true,
		},
		app:    app,
		config: config,
	}

	m.selector = dialogselector.New(app, dialogselector.Config{
		List: config,
		SelectFunc: func(row []string) tea.Cmd {
			cmd := strings.SplitN(row[0][1:], " ", 2)[1]

			var ref PageRef
			for _, p := range m.config.Pages {
				if slices.Contains(p.Commands, cmd) {
					ref = p
					break
				}
			}
			if ref.New == nil {
				return nil
			}

			return tea.Sequence(
				types.CloseActiveDialog(),
				types.OpenPage(ref.New(), true),
			)
		},
	})

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	// Re-calculate the height so the dialog is only as big as we need, up to the max
	// of the default of [DialogModel.Size].
	m.Height = min(m.Height, m.selector.GetHeight())
}

func (m *Model) GetTitle() string {
	return "Commands"
}

func (m *Model) IsCoreDialog() bool {
	return true
}

func (m *Model) Init() tea.Cmd {
	return m.selector.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height, m.Width = msg.Height, msg.Width
		cmds = append(cmds, m.selector.Update(msg))
		m.initStyles()
		return tea.Batch(cmds...)
	case styles.ThemeUpdatedMsg:
		m.initStyles()
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		cmds = append(cmds, m.selector.Update(msg))
		return tea.Batch(cmds...)
	}

	return tea.Batch(append(cmds, m.selector.Update(msg))...)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.selector.View(),
	)
}
