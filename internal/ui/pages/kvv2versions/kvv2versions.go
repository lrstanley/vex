// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package kvv2versions

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/dialogs/alert"
	"github.com/lrstanley/vex/internal/ui/dialogs/confirm"
	"github.com/lrstanley/vex/internal/ui/pages/kvviewsecret"
	"github.com/lrstanley/vex/internal/ui/styles"
)

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app           types.AppState
	mount         *types.Mount
	path          string
	latestVersion int

	// Child components.
	table *table.Model[*table.StaticRow[api.KVVersionMetadata]]
}

var columns = []*table.Column{
	{ID: "version", Title: "Version"},
	{ID: "destroyed", Title: "Destroyed"},
	{ID: "created", Title: "Created"},
	{ID: "deleted", Title: "Deleted"},
}

func New(app types.AppState, mount *types.Mount, path string) *Model {
	if mount.KVVersion() != 2 {
		panic("mount is not a KV v2 mount")
	}

	m := &Model{
		PageModel: &types.PageModel{
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
			ShortKeyBinds: []key.Binding{
				types.KeyDelete,
				types.KeyDestroy,
			},
			FullKeyBinds: [][]key.Binding{{
				types.KeyOpenEditor,
				types.KeyDelete,
				types.KeyDestroy,
			}},
		},
		app:   app,
		mount: mount,
		path:  path,
	}

	m.table = table.New(app, columns, table.Config[*table.StaticRow[api.KVVersionMetadata]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListKVv2Versions(m.UUID(), m.mount, m.path)
		},
		SelectFn: func(value *table.StaticRow[api.KVVersionMetadata]) tea.Cmd {
			return m.selectVersion(value.Value)
		},
		RowFn: func(row *table.StaticRow[api.KVVersionMetadata]) []string {
			return m.rowFn(row)
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

		switch msg := msg.Msg.(type) {
		case types.ClientListKVv2VersionsMsg:
			// Sort versions by version number (latest first)
			sort.Slice(msg.Versions, func(i, j int) bool {
				return msg.Versions[i].Version > msg.Versions[j].Version
			})

			if len(msg.Versions) > 0 {
				m.latestVersion = msg.Versions[0].Version
			}

			cmds = append(cmds, types.PageClearState())
			m.table.SetRows(table.RowsFrom(msg.Versions, func(v api.KVVersionMetadata) table.ID {
				return table.ID(strconv.Itoa(v.Version))
			}))
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, types.KeyDestroy):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.destroyVersion(v.Value)
			}
		case key.Matches(msg, types.KeyOpenEditor):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.editVersion(v.Value)
			}
		case key.Matches(msg, types.KeyDelete):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.toggleDeleteVersion(v.Value)
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) editVersion(version api.KVVersionMetadata) tea.Cmd {
	if version.Destroyed || !version.DeletionTime.IsZero() {
		return nil
	}
	return types.OpenPage(kvviewsecret.New(m.app, m.mount, m.path, version.Version, true), false)
}

func (m *Model) toggleDeleteVersion(version api.KVVersionMetadata) tea.Cmd {
	if version.Destroyed {
		return types.OpenDialog(alert.New(m.app, alert.Config{
			Title:   fmt.Sprintf("Version %d is destroyed", version.Version),
			Message: "This version has been destroyed and cannot be undeleted.",
		}))
	}
	if !version.DeletionTime.IsZero() {
		return types.OpenDialog(confirm.New(m.app, confirm.Config{
			Title:      fmt.Sprintf("Undelete Version %d", version.Version),
			Message:    "Are you sure you want to undelete this version?",
			AllowsBlur: true,
			ConfirmFn: func() tea.Cmd {
				return tea.Sequence(
					m.app.Client().UndeleteKVSecret(m.UUID(), m.mount, m.path, version.Version),
					types.CloseActiveDialog(),
					types.RefreshData(m.UUID()),
				)
			},
			CancelFn: types.CloseActiveDialog,
		}))
	}
	return types.OpenDialog(confirm.New(m.app, confirm.Config{
		Title:         fmt.Sprintf("Delete Version %d", version.Version),
		Message:       "Are you sure you want to delete this version? This can be undone.",
		AllowsBlur:    true,
		ConfirmStatus: types.Warning,
		ConfirmFn: func() tea.Cmd {
			return tea.Sequence(
				m.app.Client().DeleteKVSecret(m.UUID(), m.mount, m.path, version.Version),
				types.CloseActiveDialog(),
				types.RefreshData(m.UUID()),
			)
		},
		CancelFn: types.CloseActiveDialog,
	}))
}

func (m *Model) destroyVersion(version api.KVVersionMetadata) tea.Cmd {
	if version.Destroyed {
		return types.OpenDialog(alert.New(m.app, alert.Config{
			Title:   fmt.Sprintf("Version %d is destroyed", version.Version),
			Message: "This version has already been destroyed.",
		}))
	}
	return types.OpenDialog(confirm.New(m.app, confirm.Config{
		Title:         fmt.Sprintf("Destroy Version %d", version.Version),
		Message:       "Are you sure you want to destroy this version? This cannot be undone.",
		AllowsBlur:    true,
		ConfirmStatus: types.Error,
		ConfirmFn: func() tea.Cmd {
			return tea.Sequence(
				m.app.Client().DestroyKVSecret(m.UUID(), m.mount, m.path, version.Version),
				types.CloseActiveDialog(),
				types.RefreshData(m.UUID()),
			)
		},
		CancelFn: types.CloseActiveDialog,
	}))
}

func (m *Model) selectVersion(version api.KVVersionMetadata) tea.Cmd {
	if version.Destroyed {
		return types.OpenDialog(alert.New(m.app, alert.Config{
			Title:   fmt.Sprintf("Version %d is destroyed", version.Version),
			Message: "This version has been destroyed and cannot be restored.",
		}))
	}

	if !version.DeletionTime.IsZero() {
		return types.OpenDialog(alert.New(m.app, alert.Config{
			Title:   fmt.Sprintf("Version %d is deleted", version.Version),
			Message: "This version has been deleted. To view it, you must undelete it.",
		}))
	}

	return types.OpenPage(kvviewsecret.New(m.app, m.mount, m.path, version.Version, false), false)
}

func (m *Model) rowFn(row *table.StaticRow[api.KVVersionMetadata]) []string {
	var latest, destroyed, deleted string

	if row.Value.Version == m.latestVersion {
		latest = lipgloss.NewStyle().
			Foreground(styles.Theme.SuccessFg()).
			Bold(true).
			Render(" (latest)")
	}

	if row.Value.Destroyed {
		destroyed = lipgloss.NewStyle().Foreground(styles.Theme.ErrorFg()).Bold(true).Render("true")
	} else {
		destroyed = "false"
	}

	if row.Value.DeletionTime.IsZero() {
		deleted = "false"
	} else {
		deleted = lipgloss.NewStyle().
			Foreground(styles.Theme.ErrorFg()).
			Bold(true).
			Render(formatter.TimeRelative(row.Value.DeletionTime, true))
	}

	return []string{
		strconv.Itoa(row.Value.Version) + latest,
		destroyed,
		formatter.TimeRelative(row.Value.CreatedTime, true),
		deleted,
	}
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	return m.table.View()
}

func (m *Model) GetTitle() string {
	return m.mount.Path + m.path
}
