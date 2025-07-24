// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package aclpolicies

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/datatable"
	"github.com/lrstanley/vex/internal/ui/frames/codeframe"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands    = []string{"aclpolicies", "aclpolicy"}
	dataColumns = []string{"Name"}
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	filter string

	// Child components.
	table *datatable.Model[string]
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			ShortKeyBinds:    []key.Binding{types.KeyRefresh, types.KeyQuit},
			FullKeyBinds:     [][]key.Binding{{types.KeyRefresh, types.KeyQuit}},
		},
		app: app,
	}
	m.table = datatable.New(app, datatable.Config[string]{
		FetchFn: func(app types.AppState) tea.Cmd {
			return app.Client().ListACLPolicies(m.UUID())
		},
		SelectFn: func(app types.AppState, value string) tea.Cmd {
			return m.app.Client().GetACLPolicy(m.UUID(), value)
		},
		RowFn: func(app types.AppState, value string) []string {
			return []string{value}
		},
	})

	return m
}

func (m *Model) Init() tea.Cmd {
	return m.table.Init()
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyCancel):
			return types.ClearAppFilter()
		case key.Matches(msg, types.KeyQuit):
			return tea.Quit
		}
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
		case types.ClientListACLPoliciesMsg:
			if msg.Error == nil {
				m.table.SetData(dataColumns, vmsg.Policies)
			}
		case types.ClientGetACLPolicyMsg:
			if msg.Error == nil {
				title := "ACL Policy: " + vmsg.Name
				cmds = append(cmds, types.OpenPage(codeframe.New(m.app, title, vmsg.Content, "hcl"), false))
			}
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
	return styles.Pluralize(m.table.DataLen(), "acl policy", "acl policies")
}
