// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package mounts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/pages/secrets"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var (
	Commands    = []string{"mounts", "mount"}
	dataColumns = []string{"Path", "Type", "Description", "Accessor", "Deprecated", "Plugin Version"}
)

type Data struct {
	Mount *types.Mount
}

func (d Data) Get() Data {
	return d
}

func (d Data) Row() []string {
	var opts []string

	for k, v := range d.Mount.Options {
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
		d.Mount.Path,
		styles.Trunc(d.Mount.Type+sopts, 15),
		styles.Trunc(d.Mount.Description, 40),
		d.Mount.Accessor,
		d.Mount.DeprecationStatus,
		d.Mount.RunningVersion,
	}
}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height        int
	width         int
	filter        string
	mounts        []*types.Mount
	selectedMount *types.Mount

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
			m.selectedMount = item.Mount
		},
	})

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.app.Client().ListMounts(m.UUID()),
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
				m.app.Client().ListMounts(m.UUID()),
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
		case types.ClientListMountsMsg:
			if msg.Error == nil {
				m.mounts = vmsg.Mounts
				m.updateTableData()
			}
		}
	}

	cmds = append(cmds, m.tableComponent.Update(msg))

	if v := m.selectedMount; v != nil {
		m.selectedMount = nil
		return types.OpenPage(secrets.New(m.app, v, ""), false)
	}
	return tea.Batch(cmds...)
}

func (m *Model) updateTableData() {
	if len(m.mounts) == 0 {
		m.tableComponent.SetData([]string{}, []Data{})
		return
	}

	var mountData []Data
	for _, mount := range m.mounts {
		mountData = append(mountData, Data{Mount: mount})
	}

	m.tableComponent.SetData(
		dataColumns,
		mountData,
	)
}

func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	return m.tableComponent.View()
}

func (m *Model) TopMiddleBorder() string {
	if len(m.mounts) == 0 {
		return ""
	}
	return strconv.Itoa(len(m.mounts)) + " mounts"
}
