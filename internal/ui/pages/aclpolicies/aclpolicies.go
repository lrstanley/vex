// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package aclpolicies

import (
	"strconv"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/pages/genericcode"
)

var (
	Commands    = []string{"aclpolicies", "aclpolicy"}
	dataColumns = []string{"Name"}
)

type Data struct {
	PolicyName string
}

func (d Data) Get() Data {
	return d
}

func (d Data) Row() []string {
	return []string{
		d.PolicyName,
	}
}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height         int
	width          int
	filter         string
	policies       []string
	selectedPolicy string

	// Child components.
	tableComponent *table.Model[Data]
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
	m.tableComponent = table.New(app, table.Config[Data]{
		OnSelect: func(item Data) {
			m.selectedPolicy = item.PolicyName
		},
	})

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.app.Client().ListACLPolicies(m.UUID()),
		m.tableComponent.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyQuit):
			return tea.Quit
		case key.Matches(msg, types.KeyRefresh):
			cmds = append(
				cmds,
				m.tableComponent.SetLoading(),
				m.app.Client().ListACLPolicies(m.UUID()),
			)
		}
	case types.AppFilterMsg:
		if msg.UUID != m.UUID() {
			return nil
		}

		m.filter = msg.Text
		m.tableComponent.SetFilter(msg.Text)
	case types.ClientMsg:
		switch vmsg := msg.Msg.(type) {
		case types.ClientListACLPoliciesMsg:
			if msg.Error == nil {
				m.policies = vmsg.Policies
				m.updateTableData()
			}
		case types.ClientGetACLPolicyMsg:
			if msg.Error == nil {
				title := "ACL Policy: " + vmsg.Name
				cmds = append(cmds, types.OpenPage(genericcode.New(m.app, title, vmsg.Content, "hcl"), false))
			}
		}
	}

	cmds = append(cmds, m.tableComponent.Update(msg))

	// Handle selected policy.
	if v := m.selectedPolicy; v != "" {
		m.selectedPolicy = ""
		cmds = append(cmds, m.app.Client().GetACLPolicy(m.UUID(), v))
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateTableData() {
	if len(m.policies) == 0 {
		m.tableComponent.SetData([]string{}, []Data{})
		return
	}

	var policyData []Data
	for _, policy := range m.policies {
		policyData = append(policyData, Data{PolicyName: policy})
	}

	m.tableComponent.SetData(
		dataColumns,
		policyData,
	)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return m.tableComponent.View()
}

func (m *Model) TopMiddleBorder() string {
	if len(m.policies) == 0 {
		return ""
	}
	return strconv.Itoa(len(m.policies)) + " acl policies"
}
