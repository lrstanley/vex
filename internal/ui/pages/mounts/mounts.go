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
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/dialogs/genericcode"
	"github.com/lrstanley/vex/internal/ui/pages/secretwalker"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands = []string{"mounts", "mount"}
	columns  = []*table.Column{
		{ID: "path", Title: "Path"},
		{ID: "type", Title: "Type", MaxWidth: 15},
		{ID: "description", Title: "Description", MaxWidth: 40},
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
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
			ShortKeyBinds:    []key.Binding{types.KeyDetails},
			FullKeyBinds:     [][]key.Binding{{types.KeyDetails}},
		},
		app: app,
	}

	m.table = table.New(app, columns, table.Config[*table.StaticRow[*types.Mount]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListMounts(m.UUID(), "kv")
		},
		SelectFn: func(value *table.StaticRow[*types.Mount]) tea.Cmd {
			return types.OpenPage(secretwalker.New(app, value.Value, ""), false)
		},
		RowFn: m.rowFn,
	})

	return m
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
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyDetails):
			if v, ok := m.table.GetSelectedRow(); ok {
				return types.OpenDialog(genericcode.NewJSON(m.app, fmt.Sprintf("Mount Details: %q", v.Value.Path), v.Value))
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) rowFn(row *table.StaticRow[*types.Mount]) []string {
	var opts []string

	for k, v := range row.Value.Options {
		var name string
		switch k {
		case "version":
			name = "ver"
		}
		opts = append(opts, fmt.Sprintf("%s=%s", name, v))
	}

	sopts := strings.Join(opts, ",")
	if sopts != "" {
		sopts = " (" + sopts + ")"
	}

	return []string{
		row.Value.Path,
		row.Value.Type + sopts,
		row.Value.Description,
		row.Value.DeprecationStatus,
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
