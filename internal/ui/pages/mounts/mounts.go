// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package mounts

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/datatable"
	"github.com/lrstanley/vex/internal/ui/pages/secrets"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands    = []string{"mounts", "mount"}
	dataColumns = []string{"Path", "Type", "Description", "Accessor", "Deprecated", "Plugin Version"}
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	filter string

	// Child components.
	table *datatable.Model[*types.Mount]
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
		},
		app: app,
	}

	m.table = datatable.New(app, datatable.Config[*types.Mount]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListMounts(m.UUID())
		},
		SelectFn: func(value *types.Mount) tea.Cmd {
			return types.OpenPage(secrets.New(app, value, ""), false)
		},
		RowFn: m.rowFn,
	})

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.table.Init(),
		types.DataRefresh(m.UUID()),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case types.PageRefocusedMsg:
		return types.DataRefresh(m.UUID())
	case types.DataRefreshMsg:
		return m.table.Fetch()
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

		switch vmsg := msg.Msg.(type) {
		case types.ClientListMountsMsg:
			cmds = append(cmds, m.table.SetLoading())
			if msg.Error == nil {
				m.table.SetData(
					dataColumns,
					vmsg.Mounts,
				)
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) rowFn(value *types.Mount) []string {
	var opts []string

	for k, v := range value.Options {
		var name string
		switch k {
		case "version":
			name = "ver"
		}
		opts = append(opts, fmt.Sprintf("%s=%s", name, v))
	}

	sopts := strings.Join(opts, ",")
	if len(sopts) > 0 {
		sopts = " (" + sopts + ")"
	}

	return []string{
		value.Path,
		styles.Trunc(value.Type+sopts, 15),
		styles.Trunc(value.Description, 40),
		value.Accessor,
		value.DeprecationStatus,
		value.RunningVersion,
	}
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(m.table.DataLen(), "mount", "mounts")
}
