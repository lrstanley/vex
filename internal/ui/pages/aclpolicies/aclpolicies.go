// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package aclpolicies

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/pages/genericcode"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands = []string{"aclpolicies", "aclpolicy"}
	columns  = []*table.Column{
		{ID: "name", Title: "Name"},
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
	table *table.Model[*table.StaticRow[string]]
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
	m.table = table.New(app, columns, table.Config[*table.StaticRow[string]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListACLPolicies(m.UUID())
		},
		SelectFn: func(value *table.StaticRow[string]) tea.Cmd {
			return m.app.Client().GetACLPolicy(m.UUID(), value.Value)
		},
		RowFn: func(value *table.StaticRow[string]) []string {
			return []string{value.Value}
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
		case types.ClientListACLPoliciesMsg:
			cmds = append(cmds, types.PageClearState())
			m.table.SetRows(table.RowsFrom(vmsg.Policies, func(policy string) table.ID {
				return table.ID(policy)
			}))
		case types.ClientGetACLPolicyMsg:
			title := "ACL Policy: " + vmsg.Name
			cmds = append(cmds, types.OpenPage(genericcode.New(m.app, title, vmsg.Content, "hcl"), false))
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
	return styles.Pluralize(m.table.TotalFilteredRows(), "acl policy", "acl policies")
}
