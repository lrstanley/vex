// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package mounts

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/dialogs/genericcode"
	"github.com/lrstanley/vex/internal/ui/pages/secrets"
	"github.com/lrstanley/vex/internal/ui/pages/secretwalker"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands = []string{"mounts", "mount"}
	columns  = []*table.Column{
		{ID: "path", Title: "Path"},
		{ID: "type", Title: "Type", MaxWidth: 15},
		{ID: "description", Title: "Description", MaxWidth: 40},
		{ID: "capabilities", Title: "Capabilities"},
		{ID: "deprecated", Title: "Deprecated"},
		{ID: "plugin_version", Title: "Plugin Version"},
	}
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	filter string

	// Child components.
	table *table.Model[*table.StaticRow[*types.Mount]]

	// Styles.
	supportedFg  lipgloss.Style
	deprecatedFg lipgloss.Style
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
			ShortKeyBinds: []key.Binding{
				types.OverrideHelp(types.KeyDetails, "details"),
				types.OverrideHelp(types.KeyListRecursive, "recurse"),
			},
			FullKeyBinds: [][]key.Binding{{
				types.KeyDetails,
				types.KeyListRecursive,
			}},
		},
		app: app,
	}

	m.table = table.New(app, columns, table.Config[*table.StaticRow[*types.Mount]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListMounts(m.UUID())
		},
		SelectFn: func(value *table.StaticRow[*types.Mount]) tea.Cmd {
			return m.openMount(value)
		},
		RowFn: m.rowFn,
	})

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.supportedFg = lipgloss.NewStyle().
		Foreground(styles.Theme.SuccessFg())
	m.deprecatedFg = lipgloss.NewStyle().
		Foreground(styles.Theme.WarningFg())
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.table.Init(),
		types.RefreshData(m.UUID()),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.table.Update(msg)
	case types.PageVisibleMsg:
		return types.RefreshData(m.UUID())
	case types.RefreshDataMsg:
		return tea.Batch(
			types.PageLoading(),
			m.table.Fetch(false),
		)
	case types.AppFilterMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		m.filter = msg.Text
		m.table.SetFilter(msg.Text)
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		if msg.Error != nil {
			return types.PageErrors(msg.Error)
		}

		switch vmsg := msg.Msg.(type) {
		case types.ClientListMountsMsg:
			cmds = append(cmds, types.PageClearState())
			m.table.SetRows(table.RowsFrom(vmsg.Mounts, func(m *types.Mount) table.ID {
				return table.ID(m.Path)
			}))
		}
	case styles.ThemeUpdatedMsg:
		m.initStyles()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyDetails):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.openDetails(v)
			}
		case key.Matches(msg, types.KeyListRecursive):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.openRecursive(v)
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) openMount(row *table.StaticRow[*types.Mount]) tea.Cmd {
	if row.Value.Type != "kv" {
		return m.openDetails(row)
	}
	return types.OpenPage(secretwalker.New(m.app, row.Value, ""), false)
}

func (m *Model) openDetails(row *table.StaticRow[*types.Mount]) tea.Cmd {
	return types.OpenDialog(genericcode.NewYAML(m.app, fmt.Sprintf("Mount Details: %q", row.Value.Path), false, row.Value))
}

func (m *Model) openRecursive(row *table.StaticRow[*types.Mount]) tea.Cmd {
	return types.OpenPage(secrets.New(m.app, row.Value), false)
}

func (m *Model) rowFn(row *table.StaticRow[*types.Mount]) []string {
	var opts []string

	for k, v := range row.Value.Options {
		switch k {
		case "version":
			opts = append(opts, "v"+v)
		}
	}

	sopts := strings.Join(opts, ",")
	if sopts != "" {
		sopts = " (" + sopts + ")"
	}

	deprecationStatus := row.Value.DeprecationStatus
	if deprecationStatus == "" {
		deprecationStatus = "unknown"
	}
	if row.Value.DeprecationStatus == "supported" {
		deprecationStatus = m.supportedFg.Render(deprecationStatus)
	} else {
		deprecationStatus = m.deprecatedFg.Render(deprecationStatus)
	}

	return []string{
		row.Value.Path,
		row.Value.Type + sopts,
		row.Value.Description,
		styles.ClientCapabilities(row.Value.Capabilities, row.Value.Path),
		deprecationStatus,
		row.Value.RunningVersion,
	}
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(m.table.TotalFilteredRows(), "mount", "mounts")
}
