// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package help

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/viewport"
	"github.com/lrstanley/vex/internal/ui/styles"
	"github.com/lrstanley/x/charm/formatter"
)

const (
	colCells   = 35
	maxCols    = 3
	colPadding = 2
)

var _ types.Dialog = (*Model)(nil) // Ensure we implement the dialog interface.

type Model struct {
	*types.DialogModel

	// Core state.
	app types.AppState

	// Styles.
	keyStyle      lipgloss.Style
	keyInnerStyle lipgloss.Style
	descStyle     lipgloss.Style

	// Components.
	viewport *viewport.Model
}

func New(app types.AppState) *Model {
	m := &Model{
		DialogModel: &types.DialogModel{
			Size:            types.DialogSizeCustom,
			DisableChildren: true,
		},
		app:      app,
		viewport: viewport.New(app),
	}
	m.initStyles()
	m.Width, m.Height = m.generateHelp(0, 0)
	m.viewport.SetDimensions(m.Width, m.Height)
	return m
}

func (m *Model) initStyles() {
	m.keyStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.AppFg())

	m.keyInnerStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.ShortHelpKeyFg()).
		Bold(true)

	m.descStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.AppFg())
}

func (m *Model) GetTitle() string {
	return "Keybind Help"
}

func (m *Model) IsCoreDialog() bool {
	return true
}

func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = m.generateHelp(msg.Width, msg.Height)
		m.viewport.SetDimensions(m.Width, m.Height)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
		m.Width, m.Height = m.generateHelp(m.Width, m.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyHelp):
			return types.CloseActiveDialog()
		}
	}

	return tea.Batch(append(
		cmds,
		m.viewport.Update(msg),
	)...)
}

func (m *Model) styledKeyBinding(b key.Binding, maxKeyWidth int) string {
	return m.keyStyle.
		Width(maxKeyWidth+2). // +2 for the < and >.
		MarginRight(1).
		Render(
			m.keyStyle.Render("<")+m.keyInnerStyle.Render(b.Help().Key)+m.keyStyle.Render(">"),
		) + m.descStyle.Render(b.Help().Desc)
}

func (m *Model) generateHelp(w, h int) (resultWidth, resultHeight int) {
	if w == 0 || h == 0 {
		m.viewport.SetContent("")
		return 0, 0
	}

	var keys []key.Binding
	if m.app.Dialog().Len(true) > 0 {
		for _, k := range m.app.Dialog().FullHelp() {
			keys = append(keys, k...)
		}
	} else {
		for _, k := range m.app.Page().FullHelp() {
			keys = append(keys, k...)
		}
	}

	// grid of rows -> columns of key bindings.
	grid := [][]key.Binding{}
	cols := min(w/colCells, maxCols)
	rows := (len(keys) + cols - 1) / cols

	for i := range rows {
		grid = append(grid, []key.Binding{})
		for j := range cols {
			if i*cols+j >= len(keys) {
				break
			}
			grid[i] = append(grid[i], keys[i*cols+j])
		}
	}

	colKeyLengths := make([]int, cols) // Max key length for each column.
	for _, row := range grid {
		for i := range row {
			colKeyLengths[i] = max(colKeyLengths[i], len(row[i].Help().Key))
		}
	}

	var buf strings.Builder

	var cellContent string
	var cellWidth int
	for _, row := range grid {
		for i := range row {
			cellContent = m.styledKeyBinding(row[i], colKeyLengths[i])
			cellWidth = ansi.StringWidth(cellContent)

			if cellWidth < colCells-colPadding {
				cellContent += strings.Repeat(" ", colCells-colPadding-cellWidth)
			} else if cellWidth > colCells-colPadding {
				cellContent = formatter.Trunc(cellContent, colCells-colPadding)
			}

			buf.WriteString(cellContent)
		}
		buf.WriteString("\n")
	}

	m.viewport.SetContent(strings.TrimSuffix(buf.String(), "\n"))
	return cols*colCells + m.viewport.Styles().Base.GetHorizontalPadding(), min(rows, h)
}

func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}
	return m.viewport.View()
}
