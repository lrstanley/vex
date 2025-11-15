// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package kvviewsecret

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/confirmable"
	"github.com/lrstanley/vex/internal/ui/components/viewport"
	"github.com/lrstanley/vex/internal/ui/dialogs/confirm"
	"github.com/lrstanley/vex/internal/ui/dialogs/genericcode"
	"github.com/lrstanley/vex/internal/ui/dialogs/textarea"
	"github.com/lrstanley/vex/internal/ui/styles"
	"github.com/lrstanley/x/charm/formatter"
)

// TODO: change edit and delete to always be staged. i.e. edit (or delete) multiple things, then apply.

type item struct {
	model  *Model
	key    string
	value  any
	masked bool
}

func (i *item) ValueString() string {
	return fmt.Sprintf("%v", i.value)
}

func (i *item) IsMultiLine() bool {
	v := i.ValueString()
	return styles.W(v) > i.model.width || styles.H(v) > 1
}

func (i *item) FilterValue() string {
	return i.key
}

func (i *item) Title() string {
	if slices.Contains(i.model.keysMarkedForDeletion, i.key) {
		return lipgloss.NewStyle().
			Foreground(styles.Theme.ErrorFg()).
			Render(i.key)
	}
	return i.key
}

func (i *item) Description() string {
	var styleFunc func(...string) string

	switch {
	case slices.Contains(i.model.keysMarkedForDeletion, i.key):
		styleFunc = lipgloss.NewStyle().Foreground(styles.Theme.ErrorFg()).Bold(true).Render
	case !i.masked:
		styleFunc = lipgloss.NewStyle().Foreground(styles.Theme.AppFg()).Render
	default:
		styleFunc = func(v ...string) string {
			return strings.Join(v, " ")
		}
	}

	if i.IsMultiLine() {
		return styleFunc("<press 'x' to view full value>")
	}

	if i.masked {
		return styleFunc(formatter.MaskReplacementValue)
	}

	return styleFunc(i.ValueString())
}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height                int
	width                 int
	openedAsEditor        bool
	mount                 *types.Mount
	path                  string
	version               int
	data                  map[string]any
	filter                string
	isFlat                bool
	forceJSON             bool
	isNonFlatMasked       bool // If masking is enabled for non-flat (json) masking.
	unmaskedKeys          []string
	keysMarkedForDeletion []string // Keys marked for deletion, flat-only.

	// Child components.
	delegate list.DefaultDelegate
	list     list.Model
	viewport *viewport.Model
}

func New(app types.AppState, mount *types.Mount, path string, version int, openedAsEditor bool) *Model {
	// TODO:
	//   - use scrollbar vs paginator. maybe if we implement a custom component?
	m := &Model{
		PageModel: &types.PageModel{
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
			ShortKeyBinds: []key.Binding{
				types.OverrideHelp(types.KeySelectItem, "edit"),
				types.KeyCopy,
				types.KeyToggleMask,
			},
			FullKeyBinds: [][]key.Binding{{
				types.OverrideHelp(types.KeySelectItem, "edit"),
				types.KeyOpenEditor,
				types.KeyCopy,
				types.KeyToggleMask,
				types.KeyToggleMaskAll,
				types.KeyRenderJSON,
				types.KeyToggleDelete,
				types.KeyDelete,
			}},
		},
		app:             app,
		openedAsEditor:  openedAsEditor,
		mount:           mount,
		path:            path,
		version:         version,
		isNonFlatMasked: true,
		viewport:        viewport.New(app),
	}
	m.delegate = list.NewDefaultDelegate()
	m.list = list.New(nil, m.delegate, 0, 0)

	m.list.SetFilteringEnabled(true)
	m.list.SetShowFilter(false)
	m.list.SetShowHelp(false)
	m.list.SetShowPagination(true)
	m.list.SetShowStatusBar(true)
	m.list.SetShowTitle(false)
	m.list.DisableQuitKeybindings()
	m.list.SetStatusBarItemName("secret", "secrets")

	m.setStyle()
	return m
}

func (m *Model) setStyle() {
	m.list.Styles.ActivePaginationDot = m.list.Styles.ActivePaginationDot.
		Foreground(styles.Theme.ScrollbarThumbFg())

	m.list.Styles.InactivePaginationDot = m.list.Styles.InactivePaginationDot.
		Foreground(styles.Theme.ScrollbarTrackFg())

	m.list.Styles.NoItems = m.list.Styles.NoItems.
		Foreground(styles.Theme.AppFg()).
		PaddingLeft(2)

	m.list.Styles.StatusBar = m.list.Styles.StatusBar.
		Foreground(styles.Theme.AppFg())

	m.list.Styles.StatusEmpty = m.list.Styles.StatusEmpty.
		Foreground(styles.Theme.AppFg())

	m.list.Styles.NoItems = m.list.Styles.NoItems.
		Foreground(styles.Theme.AppFg())

	m.list.Styles.StatusBarFilterCount = m.list.Styles.StatusBarFilterCount.
		Foreground(styles.Theme.AppFg()).
		Faint(true)

	m.list.Styles.StatusBarActiveFilter = m.list.Styles.StatusBarActiveFilter.
		Foreground(styles.Theme.AppFg()).
		Faint(true)

	m.list.Styles.DividerDot = m.list.Styles.DividerDot.
		Foreground(styles.Theme.AppFg()).
		Faint(true)

	delStyles := list.NewDefaultItemStyles(true)

	delStyles.NormalTitle = delStyles.NormalTitle.
		Foreground(styles.Theme.ListItemFg())

	delStyles.NormalDesc = delStyles.NormalDesc.
		Foreground(styles.Theme.ListItemFg())

	delStyles.SelectedTitle = delStyles.SelectedTitle.
		Foreground(styles.Theme.ListItemSelectedFg()).
		BorderLeftForeground(styles.Theme.ListItemSelectedFg()).
		Bold(true)

	delStyles.SelectedDesc = delStyles.SelectedDesc.
		Foreground(styles.Theme.ListItemSelectedFg()).
		BorderLeftForeground(styles.Theme.ListItemSelectedFg())

	delStyles.DimmedTitle = delStyles.DimmedTitle.
		Foreground(styles.Theme.ListItemFg()).
		Faint(true)

	m.delegate.Styles = delStyles
	m.list.SetDelegate(m.delegate)
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.viewport.Init(),
		types.RefreshData(m.UUID()),
	)
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.list.SetSize(m.width, m.height)
		m.viewport.SetDimensions(m.width, m.height)
		return nil
	case types.PageVisibleMsg:
		return types.RefreshData(m.UUID())
	case types.RefreshDataMsg:
		return tea.Batch(
			types.PageLoading(),
			m.app.Client().GetKVSecret(m.UUID(), m.mount, m.path, m.version),
		)
	case types.AppFilterMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		m.filter = msg.Text
		if m.filter != "" {
			m.list.SetFilterText(msg.Text)
		} else {
			m.list.ResetFilter()
		}
	case types.ClientMsg:
		if msg.UUID != m.UUID() {
			return nil
		}
		if msg.Error != nil {
			return types.PageErrors(msg.Error)
		}

		switch vmsg := msg.Msg.(type) {
		case types.ClientGetSecretMsg:
			m.data = vmsg.Data
			m.isFlat = formatter.IsFlatValue(vmsg.Data)
			m.SupportFiltering = m.isFlat

			return tea.Batch(append(
				cmds,
				m.setFromData(),
				types.PageClearState(),
			)...)
		case types.ClientSuccessMsg:
			return tea.Batch(
				types.SendStatus(vmsg.Message, types.Success, 2*time.Second),
				types.RefreshData(m.UUID()),
			)
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg.Key(), types.KeyCopy):
			if !m.isFlat || m.forceJSON {
				b, err := json.MarshalIndent(m.data, "", "    ")
				if err == nil {
					return types.SetClipboard(string(b))
				}
			} else {
				item := m.getSelectedItem()
				if item == nil {
					return nil
				}
				return types.SetClipboard(item.ValueString())
			}
		case key.Matches(msg.Key(), types.KeyToggleMask, types.KeyToggleMaskAll):
			return m.toggleMasking(key.Matches(msg.Key(), types.KeyToggleMaskAll))
		case key.Matches(msg.Key(), types.KeyRenderJSON):
			m.forceJSON = !m.forceJSON
			return m.setFromData()
		case key.Matches(msg.Key(), types.KeySelectItem):
			return m.edit()
		case key.Matches(msg.Key(), types.KeyOpenEditor):
			return m.editWithEditor()
		case key.Matches(msg.Key(), types.KeyDelete):
			return m.delete()
		case msg.String() == "d" && m.isFlat && !m.forceJSON:
			return m.toggleKeyDeletion()
		}
	case styles.ThemeUpdatedMsg:
		m.setStyle()
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd, m.viewport.Update(msg))
		return tea.Batch(cmds...)
	}

	if m.isFlat && !m.forceJSON {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		cmds = append(cmds, m.viewport.Update(msg))
	}

	return tea.Batch(cmds...)
}

func (m *Model) getSelectedItem() *item {
	if !m.isFlat || m.forceJSON {
		return nil
	}
	item, ok := m.list.Items()[m.list.GlobalIndex()].(*item)
	if !ok {
		return nil
	}
	return item
}

func (m *Model) setFromData() tea.Cmd {
	var cmds []tea.Cmd

	if !m.isFlat || m.forceJSON {
		m.viewport.SetCode(formatter.ToJSON(m.data, m.isNonFlatMasked, 2), "json")
	} else {
		// Preserve marked keys, but remove any that no longer exist in the data
		if len(m.keysMarkedForDeletion) > 0 {
			m.keysMarkedForDeletion = slices.DeleteFunc(m.keysMarkedForDeletion, func(k string) bool {
				_, exists := m.data[k]
				return !exists
			})
		}

		var values []list.Item
		keys := slices.Collect(maps.Keys(m.data))
		slices.Sort(keys)
		for _, k := range keys {
			values = append(values, &item{
				model:  m,
				key:    k,
				value:  m.data[k],
				masked: !slices.Contains(m.unmaskedKeys, k),
			})
		}
		cmds = append(cmds, m.list.SetItems(values))
	}

	if m.openedAsEditor {
		m.openedAsEditor = false
		cmds = append(cmds, m.editWithEditor())
	}

	return tea.Batch(cmds...)
}

func (m *Model) items() (out []*item) {
	for _, v := range m.list.Items() {
		out = append(out, v.(*item)) //nolint:errcheck
	}
	return out
}

func (m *Model) toggleMasking(global bool) tea.Cmd {
	switch {
	case m.isFlat && !m.forceJSON && global:
		var unmaskable []*item
		for _, item := range m.items() {
			if !item.IsMultiLine() {
				unmaskable = append(unmaskable, item)
			}
		}

		if len(m.unmaskedKeys) != len(unmaskable) {
			m.unmaskedKeys = nil
			for _, item := range unmaskable {
				m.unmaskedKeys = append(m.unmaskedKeys, item.key)
			}
		} else {
			m.unmaskedKeys = []string{}
		}
	case m.isFlat && !m.forceJSON:
		item := m.getSelectedItem()
		if item == nil {
			return nil
		}
		if item.IsMultiLine() {
			return types.OpenDialog(genericcode.New(m.app, fmt.Sprintf("Secret key value: %q", item.key), item.ValueString(), ""))
		}
		if slices.Contains(m.unmaskedKeys, item.key) {
			m.unmaskedKeys = slices.DeleteFunc(m.unmaskedKeys, func(k string) bool {
				return k == item.key
			})
		} else {
			m.unmaskedKeys = append(m.unmaskedKeys, item.key)
		}

		return m.setFromData()
	default:
		m.isNonFlatMasked = !m.isNonFlatMasked
	}
	return tea.Batch(
		m.setFromData(),
		types.SendStatus("masking toggled", types.Info, 1*time.Second),
	)
}

func (m *Model) edit() tea.Cmd {
	if !m.isFlat {
		return types.OpenDialog(textarea.New(
			m.app,
			confirmable.Config[string]{
				CancelText:  "cancel",
				ConfirmText: "save",
				ConfirmFn: func(v string) tea.Cmd {
					data := map[string]any{}
					err := json.Unmarshal([]byte(v), &data)
					if err != nil {
						return types.SendStatus("failed to unmarshal json", types.Error, 2*time.Second)
					}
					return m.app.Client().PutKVSecret(m.UUID(), m.mount, m.path, data)
				},
				PassthroughTab: true,
			},
			"Edit secret",
			formatter.ToJSON(m.data, false, 2),
		))
	}

	item := m.getSelectedItem()
	if item == nil {
		return nil
	}
	return types.OpenDialog(textarea.New(
		m.app,
		confirmable.Config[string]{
			CancelText:  "cancel",
			ConfirmText: "save",
			ConfirmFn: func(v string) tea.Cmd {
				data := maps.Clone(m.data)
				data[item.key] = strings.TrimSuffix(v, "\n")
				return m.app.Client().PutKVSecret(m.UUID(), m.mount, m.path, data)
			},
			PassthroughTab: item.IsMultiLine(),
		},
		fmt.Sprintf("Edit key: %q", item.key),
		item.ValueString(),
	))
}

func (m *Model) editWithEditor() tea.Cmd {
	if m.data == nil {
		return nil
	}

	if !m.isFlat {
		return types.OpenTempEditor(
			m.UUID(),
			"update-secret-*.json",
			formatter.ToJSON(m.data, false, 2),
			func(msg types.EditorResultMsg) tea.Cmd {
				if !msg.HasChanged {
					return types.SendStatus("no changes detected", types.Info, 2*time.Second)
				}
				data := map[string]any{}
				err := json.Unmarshal([]byte(msg.After), &data)
				if err != nil {
					return types.SendStatus("failed to unmarshal json", types.Error, 2*time.Second)
				}
				return m.app.Client().PutKVSecret(m.UUID(), m.mount, m.path, data)
			},
		)
	}

	item := m.getSelectedItem()
	if item == nil {
		return nil
	}

	return types.OpenTempEditor(
		m.UUID(),
		"update-secret-*.json",
		item.ValueString(),
		func(msg types.EditorResultMsg) tea.Cmd {
			if !msg.HasChanged {
				return types.SendStatus("no changes detected", types.Info, 2*time.Second)
			}
			data := maps.Clone(m.data)
			data[item.key] = strings.TrimSuffix(msg.After, "\n")
			return m.app.Client().PutKVSecret(m.UUID(), m.mount, m.path, data)
		},
	)
}

func (m *Model) View() string {
	if !m.isFlat || m.forceJSON {
		return m.viewport.View()
	}
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(m.list.View())
}

func (m *Model) GetTitle() string {
	return m.mount.Path + m.path
}

func (m *Model) toggleKeyDeletion() tea.Cmd {
	if !m.isFlat || m.forceJSON {
		return nil
	}

	item := m.getSelectedItem()
	if item == nil {
		return nil
	}

	if slices.Contains(m.keysMarkedForDeletion, item.key) {
		m.keysMarkedForDeletion = slices.DeleteFunc(m.keysMarkedForDeletion, func(k string) bool {
			return k == item.key
		})
	} else {
		m.keysMarkedForDeletion = append(m.keysMarkedForDeletion, item.key)
	}

	return m.setFromData()
}

func (m *Model) delete() tea.Cmd {
	var title, message string
	var confirmFn func() tea.Cmd

	var all bool

	switch {
	case m.isFlat && !m.forceJSON:
		if len(m.keysMarkedForDeletion) > 0 && len(m.keysMarkedForDeletion) < len(slices.Collect(maps.Keys(m.data))) {
			var keys []string

			for _, key := range m.keysMarkedForDeletion {
				keys = append(keys, lipgloss.NewStyle().Foreground(styles.Theme.ErrorFg()).Bold(true).Render(styles.IconCaution()+" "+key))
			}

			title = fmt.Sprintf("Delete keys from %s", m.mount.Path+m.path)
			message = fmt.Sprintf(
				"Are you sure you want to delete the following keys?:\n%s\n\nThis cannot be undone.",
				strings.Join(keys, "\n"),
			)
			confirmFn = func() tea.Cmd {
				data := maps.Clone(m.data)
				for _, key := range m.keysMarkedForDeletion {
					delete(data, key)
				}
				return tea.Sequence(
					m.app.Client().PutKVSecret(m.UUID(), m.mount, m.path, data),
					types.CloseActiveDialog(),
				)
			}
		} else {
			all = true
		}
	case !m.isFlat:
		all = true
	}

	if all {
		title = fmt.Sprintf("Delete %s", m.mount.Path+m.path)
		message = "Are you sure you want to delete this secret? This cannot be undone."
		confirmFn = func() tea.Cmd {
			return tea.Sequence(
				m.app.Client().DeleteKVSecret(m.UUID(), m.mount, m.path),
				types.CloseActiveDialog(),
				types.CloseActivePage(),
			)
		}
	}

	if title == "" || message == "" || confirmFn == nil {
		return nil
	}

	return types.OpenDialog(confirm.New(m.app, confirm.Config{
		Title:         title,
		Message:       message,
		AllowsBlur:    true,
		ConfirmStatus: types.Error,
		ConfirmFn:     confirmFn,
		CancelFn:      types.CloseActiveDialog,
	}))
}

func (m *Model) TopRightBorder() string {
	var hasUnmasked bool
	if m.isFlat && !m.forceJSON {
		hasUnmasked = len(m.unmaskedKeys) > 0
	} else {
		hasUnmasked = !m.isNonFlatMasked
	}

	if hasUnmasked {
		return lipgloss.NewStyle().
			Foreground(styles.Theme.ErrorFg()).
			Background(styles.Theme.ErrorBg()).
			Padding(0, 1).
			Render(styles.IconCaution() + " unmasked secrets")
	}
	return ""
}
