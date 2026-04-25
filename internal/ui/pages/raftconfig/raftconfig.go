// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in the
// LICENSE file.

package raftconfig

import (
	"fmt"
	"strconv"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/dialogs/confirm"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var Commands = []string{"raftconfig"}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	app types.AppState

	table *table.Model[*table.StaticRow[*types.RaftConfigPeer]]
}

func New(app types.AppState) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
			ShortKeyBinds: []key.Binding{
				types.OverrideHelp(types.KeyDelete, "remove peer"),
			},
			FullKeyBinds: [][]key.Binding{{
				types.KeyDelete,
			}},
		},
		app: app,
	}

	m.table = table.New(app, table.Config[*table.StaticRow[*types.RaftConfigPeer]]{
		Columns: []*table.Column[*table.StaticRow[*types.RaftConfigPeer]]{
			{
				ID:    "node_id",
				Title: "Node ID",
				AccessorFn: func(row *table.StaticRow[*types.RaftConfigPeer]) string {
					return row.Value.NodeID
				},
			},
			{
				ID:    "address",
				Title: "Address",
				AccessorFn: func(row *table.StaticRow[*types.RaftConfigPeer]) string {
					return row.Value.Address
				},
			},
			{
				ID:       "leader",
				Title:    "Leader",
				MaxWidth: 8,
				AccessorFn: func(row *table.StaticRow[*types.RaftConfigPeer]) string {
					return strconv.FormatBool(row.Value.Leader)
				},
				StyleFn: func(row *table.StaticRow[*types.RaftConfigPeer], baseStyle lipgloss.Style, _, _ bool) lipgloss.Style {
					if row.Value.Leader {
						return baseStyle.Foreground(styles.Theme.SuccessFg())
					}
					return baseStyle
				},
			},
			{
				ID:       "voter",
				Title:    "Voter",
				MaxWidth: 8,
				AccessorFn: func(row *table.StaticRow[*types.RaftConfigPeer]) string {
					return strconv.FormatBool(row.Value.Voter)
				},
			},
			{
				ID:       "protocol_version",
				Title:    "Protocol Version",
				MaxWidth: 16,
				AccessorFn: func(row *table.StaticRow[*types.RaftConfigPeer]) string {
					return formatProtocolVersion(row.Value.ProtocolVersion)
				},
			},
		},
		FetchFn: func() tea.Cmd {
			return app.Client().GetRaftConfig(m.UUID())
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
		m.table.SetFilter(msg.Text)
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		if msg.Error != nil {
			return types.PageErrors(msg.Error)
		}

		switch vmsg := msg.Msg.(type) {
		case types.ClientRaftConfigMsg:
			cmds = append(cmds, types.PageClearState())
			m.table.SetRows(table.RowsFrom(vmsg.Peers, func(p *types.RaftConfigPeer) table.ID {
				if p.NodeID == "" {
					return table.ID(p.Address)
				}
				return table.ID(p.NodeID)
			}))
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyDelete):
			if row, ok := m.table.GetSelectedRow(); ok {
				return m.removePeer(row.Value)
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) removePeer(peer *types.RaftConfigPeer) tea.Cmd {
	if peer == nil {
		return nil
	}
	return types.OpenDialog(confirm.New(m.app, confirm.Config{
		Title:         "Remove raft peer",
		Message:       fmt.Sprintf("Remove peer %q (node_id %q)?", peer.Address, peer.NodeID),
		AllowsBlur:    true,
		ConfirmStatus: types.Warning,
		ConfirmFn: func() tea.Cmd {
			return tea.Sequence(
				m.app.Client().RemoveRaftPeer(m.UUID(), peer.NodeID),
				types.CloseActiveDialog(),
				types.RefreshData(m.UUID()),
			)
		},
		CancelFn: types.CloseActiveDialog,
	}))
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func formatProtocolVersion(version string) string {
	if len(version) == 1 && version[0] < 10 {
		return strconv.Itoa(int(version[0]))
	}
	return version
}

func (m *Model) TopMiddleBorder() string {
	return styles.Pluralize(m.table.TotalFilteredRows(), "peer", "peers")
}
