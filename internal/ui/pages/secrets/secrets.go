// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package secrets

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/pages/viewsecret"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type Data struct {
	Mount        *types.Mount
	Path         string
	FullPath     string
	Capabilities types.ClientCapabilities
	Incomplete   bool
}

var (
	Commands = []string{"secrets", "secret"}
	columns  = []*table.Column{
		{ID: "full_path", Title: "Full Path"},
		{ID: "mount_type", Title: "Mount Type"},
		{ID: "capabilities", Title: "Capabilities"},
	}
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	filter string
	data   *types.ClientListAllSecretsRecursiveMsg

	// Child components.
	table *table.Model[*table.StaticRow[*Data]]
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			RefreshInterval:  60 * time.Second,
		},
		app: app,
	}

	m.table = table.New(app, columns, table.Config[*table.StaticRow[*Data]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListAllSecretsRecursive(m.UUID())
		},
		SelectFn: func(value *table.StaticRow[*Data]) tea.Cmd {
			return types.OpenPage(viewsecret.New(m.app, value.Value.Mount, value.Value.Path), false)
		},
		RowFn: func(value *table.StaticRow[*Data]) []string {
			return []string{
				value.Value.FullPath,
				value.Value.Mount.Type,
				styles.ClientCapabilities(value.Value.Capabilities, value.Value.FullPath),
			}
		},
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
		case types.ClientListAllSecretsRecursiveMsg:
			cmds = append(cmds, types.PageClearState())

			m.data = &vmsg
			var data []*Data

			for ref := range vmsg.Tree.IterRefs() {
				if !ref.IsSecret() {
					continue
				}

				data = append(data, &Data{
					Mount:        ref.Mount,
					Path:         ref.Path,
					FullPath:     ref.GetFullPath(),
					Capabilities: ref.Capabilities,
					Incomplete:   ref.Incomplete,
				})
			}

			m.table.SetRows(table.RowsFrom(data, func(d *Data) table.ID {
				return table.ID(d.FullPath)
			}))
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(m.table.TotalFilteredRows(), "secret", "secrets")
}

func (m *Model) TopRightBorder() string {
	if m.data == nil {
		return ""
	}
	return fmt.Sprintf("requests: %d/%d", m.data.Requests, m.data.MaxRequests)
}
