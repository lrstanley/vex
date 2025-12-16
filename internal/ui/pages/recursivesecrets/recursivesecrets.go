// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package recursivesecrets

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/table"
	"github.com/lrstanley/vex/internal/ui/dialogs/confirm"
	"github.com/lrstanley/vex/internal/ui/dialogs/genericcode"
	"github.com/lrstanley/vex/internal/ui/pages/kvv2versions"
	"github.com/lrstanley/vex/internal/ui/pages/kvviewsecret"
	"github.com/lrstanley/vex/internal/ui/styles"
	"github.com/lrstanley/x/charm/formatter"
)

var Commands = []string{"secrets", "secret"}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app    types.AppState
	width  int
	height int

	// UI state.
	filter string
	mount  *types.Mount // Optional mount to restrict to.
	data   *types.ClientListAllSecretsRecursiveMsg

	// Styles.
	tooManyRequestsStyle lipgloss.Style

	// Child components.
	table *table.Model[*table.StaticRow[*types.ClientSecretTreeRef]]
}

func New(app types.AppState, mount *types.Mount) *Model {
	m := &Model{
		PageModel: &types.PageModel{
			Commands:         Commands,
			SupportFiltering: true,
			RefreshInterval:  60 * time.Second,
			ShortKeyBinds: []key.Binding{
				types.OverrideHelp(types.KeyDetails, "details"),
				types.KeyDelete,
			},
			FullKeyBinds: [][]key.Binding{{
				types.KeyDetails,
				types.KeyOpenEditor,
				types.KeyDelete,
			}},
		},
		app:   app,
		mount: mount,
	}

	columns := []*table.Column{
		{ID: "full_path", Title: "Full Path"},
		{ID: "mount_type", Title: "Mount Type"},
		{ID: "capabilities", Title: "Capabilities"},
	}

	m.table = table.New(app, columns, table.Config[*table.StaticRow[*types.ClientSecretTreeRef]]{
		FetchFn: func() tea.Cmd {
			return app.Client().ListAllSecretsRecursive(m.UUID(), m.mount)
		},
		SelectFn: func(value *table.StaticRow[*types.ClientSecretTreeRef]) tea.Cmd {
			return m.selectSecret(value.Value)
		},
		RowFn: func(value *table.StaticRow[*types.ClientSecretTreeRef]) []string {
			return []string{
				styles.IconSecret() + " " + value.Value.GetFullPath(true),
				value.Value.Mount.Type,
				styles.ClientCapabilities(value.Value.Capabilities, value.Value.GetFullPath(true)),
			}
		},
	})

	m.initStyles()
	return m
}

func (m *Model) initStyles() {
	m.tooManyRequestsStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.ErrorFg()).
		Background(styles.Theme.ErrorBg()).
		Padding(0, 1).
		Align(lipgloss.Center)
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
		m.setDimensions(msg.Width, msg.Height)
		return nil
	case styles.ThemeUpdatedMsg:
		m.initStyles()
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
			var data []*types.ClientSecretTreeRef

			for ref := range vmsg.Tree.IterRefs() {
				if !ref.IsSecret() {
					continue
				}
				data = append(data, ref)
			}

			slices.SortFunc(data, func(a, b *types.ClientSecretTreeRef) int {
				return strings.Compare(a.GetFullPath(true), b.GetFullPath(true))
			})

			m.table.SetRows(table.RowsFrom(data, func(d *types.ClientSecretTreeRef) table.ID {
				return table.ID(d.GetFullPath(true))
			}))

			m.setDimensions(m.width, m.height)
		case types.ClientGetKVv2MetadataMsg:
			return types.OpenDialog(genericcode.NewYAML(
				m.app,
				"Metadata: "+vmsg.Mount.Path+vmsg.Path,
				false,
				vmsg.Metadata,
			))
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.KeyDetails):
			if v, ok := m.table.GetSelectedRow(); ok && !strings.HasSuffix(v.Value.Path, "/") && v.Value.Mount.KVVersion() == 2 {
				return m.app.Client().GetKVv2Metadata(m.UUID(), v.Value.Mount, v.Value.Path)
			}
		case key.Matches(msg, types.KeyOpenEditor):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.editSecret(v.Value)
			}
		case key.Matches(msg, types.KeyDelete):
			if v, ok := m.table.GetSelectedRow(); ok {
				return m.deleteSecret(v.Value)
			}
		}
	}

	return tea.Batch(append(cmds, m.table.Update(msg))...)
}

func (m *Model) setDimensions(width, height int) {
	m.width = width
	m.height = height
	if m.data != nil && m.data.RequestAttempts > m.data.MaxRequests {
		m.table.SetDimensions(m.width, m.height-1)
	} else {
		m.table.SetDimensions(m.width, m.height)
	}
}

func (m *Model) selectSecret(secret *types.ClientSecretTreeRef) tea.Cmd {
	if secret.Mount.KVVersion() == 2 {
		return types.OpenPage(kvv2versions.New(m.app, secret.Mount, secret.GetFullPath(false)), false)
	}
	return types.OpenPage(kvviewsecret.New(m.app, secret.Mount, secret.GetFullPath(false), 0, false), false)
}

func (m *Model) editSecret(secret *types.ClientSecretTreeRef) tea.Cmd {
	if strings.HasSuffix(secret.Path, "/") {
		return nil
	}
	return types.OpenPage(kvviewsecret.New(m.app, secret.Mount, secret.GetFullPath(false), 0, true), false)
}

func (m *Model) deleteSecret(secret *types.ClientSecretTreeRef) tea.Cmd {
	if strings.HasSuffix(secret.Path, "/") {
		return nil
	}

	var confirmCmds []tea.Cmd
	if len(m.table.GetAllRows()) > 1 {
		confirmCmds = []tea.Cmd{
			m.app.Client().DeleteKVSecret(m.UUID(), secret.Mount, secret.GetFullPath(false)),
			types.CloseActiveDialog(),
			types.RefreshData(m.UUID()),
		}
	} else {
		confirmCmds = []tea.Cmd{
			m.app.Client().DeleteKVSecret(m.UUID(), secret.Mount, secret.GetFullPath(false)),
			types.CloseActiveDialog(),
			types.CloseActivePage(),
		}
	}

	return types.OpenDialog(confirm.New(m.app, confirm.Config{
		Title:         fmt.Sprintf("Delete %s", secret.GetFullPath(true)),
		Message:       "Are you sure you want to delete this secret? This cannot be undone.",
		AllowsBlur:    true,
		ConfirmStatus: types.Error,
		ConfirmFn: func() tea.Cmd {
			return tea.Sequence(confirmCmds...)
		},
		CancelFn: types.CloseActiveDialog,
	}))
}

func (m *Model) View() string {
	if m.table.Width == 0 || m.table.Height == 0 {
		return ""
	}
	if m.data != nil && m.data.RequestAttempts > m.data.MaxRequests {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.tooManyRequestsStyle.
				Width(m.width).
				Render(formatter.Trunc(fmt.Sprintf(
					"hit request limit trying to list: %d/%d (need at least %d)",
					m.data.Requests,
					m.data.MaxRequests,
					m.data.RequestAttempts,
				), m.width-m.tooManyRequestsStyle.GetHorizontalFrameSize())),
			m.table.View(),
		)
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

func (m *Model) GetTitle() string {
	if m.mount == nil {
		return "Secrets (recursive)"
	}

	return fmt.Sprintf("Secrets (recursive): %s", m.mount.Path)
}
