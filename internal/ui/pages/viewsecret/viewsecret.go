// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package viewsecret

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/viewport"
	"github.com/lrstanley/vex/internal/ui/dialogs/genericcode"
	"github.com/lrstanley/vex/internal/ui/styles"
)

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
	v := strings.TrimSpace(i.ValueString())
	return styles.W(v) > i.model.width || styles.H(v) > 1
}

func (i *item) FilterValue() string {
	return i.key
}

func (i *item) Title() string {
	return i.key
}

func (i *item) Description() string {
	if i.IsMultiLine() {
		return "<truncated, use 'd' to view full value>"
	}

	if i.masked {
		return formatter.MaskReplacementValue
	}

	svalue := i.ValueString()
	return lipgloss.NewStyle().
		Foreground(styles.Theme.Fg()).
		Render(svalue)
}

var _ types.Page = (*Model)(nil) // Ensure we implement the page interface.

type Model struct {
	*types.PageModel

	// Core state.
	app types.AppState

	// UI state.
	height          int
	width           int
	mount           *types.Mount
	path            string
	data            map[string]any
	filter          string
	isFlat          bool
	forceJSON       bool
	isNonFlatMasked bool // If masking is enabled for non-flat (json) masking.
	unmaskedKeys    []string

	// Child components.
	delegate list.DefaultDelegate
	list     list.Model
	viewport *viewport.Model
}

func New(app types.AppState, mount *types.Mount, path string) *Model {
	// TODO:
	//   - use scrollbar vs paginator. maybe if we implement a custom component?
	m := &Model{
		PageModel: &types.PageModel{
			SupportFiltering: true,
			RefreshInterval:  30 * time.Second,
			ShortKeyBinds:    []key.Binding{},
			FullKeyBinds: [][]key.Binding{{
				types.KeyCopy,
				types.KeyToggleMask,
				types.KeyToggleMaskAll,
				types.KeyRenderJSON,
			}},
		},
		app:             app,
		mount:           mount,
		path:            path,
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
		Foreground(styles.Theme.Fg()).
		PaddingLeft(2)

	m.list.Styles.StatusBar = m.list.Styles.StatusBar.
		Foreground(styles.Theme.Fg())

	m.list.Styles.StatusEmpty = m.list.Styles.StatusEmpty.
		Foreground(styles.Theme.Fg())

	m.list.Styles.NoItems = m.list.Styles.NoItems.
		Foreground(styles.Theme.Fg())

	m.list.Styles.StatusBarFilterCount = m.list.Styles.StatusBarFilterCount.
		Foreground(styles.Theme.Fg()).
		Faint(true)

	m.list.Styles.StatusBarActiveFilter = m.list.Styles.StatusBarActiveFilter.
		Foreground(styles.Theme.Fg()).
		Faint(true)

	m.list.Styles.DividerDot = m.list.Styles.DividerDot.
		Foreground(styles.Theme.Fg()).
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
		m.viewport.SetHeight(m.height)
		m.viewport.SetWidth(m.width)
		return nil
	case types.PageVisibleMsg:
		return types.RefreshData(m.UUID())
	case types.RefreshDataMsg:
		return tea.Batch(
			types.PageLoading(),
			m.app.Client().GetSecret(m.UUID(), m.mount, m.path),
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
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg.Key(), types.KeyCopy):
			if !m.isFlat {
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
		case key.Matches(msg.Key(), types.KeyDetails):
			item := m.getSelectedItem()
			if item == nil {
				return nil
			}
			if !item.IsMultiLine() {
				return nil
			}
			return types.OpenDialog(genericcode.New(m.app, fmt.Sprintf("Secret key value: %q", item.key), item.ValueString(), ""))
		case key.Matches(msg.Key(), types.KeyRenderJSON):
			m.forceJSON = !m.forceJSON
			return m.setFromData()
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
	item, ok := m.list.Items()[m.list.GlobalIndex()].(*item)
	if !ok {
		return nil
	}
	return item
}

func (m *Model) setFromData() tea.Cmd {
	if !m.isFlat || m.forceJSON {
		m.viewport.SetCode(formatter.ToJSON(m.data, m.isNonFlatMasked, 2), "json")
		return nil
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
	return m.list.SetItems(values)
}

func (m *Model) toggleMasking(global bool) tea.Cmd {
	switch {
	case m.isFlat && !m.forceJSON && global:
		if len(m.unmaskedKeys) != len(m.data) {
			m.unmaskedKeys = slices.Collect(maps.Keys(m.data))
		} else {
			m.unmaskedKeys = []string{}
		}
	case m.isFlat && !m.forceJSON:
		item, ok := m.list.Items()[m.list.GlobalIndex()].(*item)
		if !ok {
			return nil
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
	return fmt.Sprintf("Viewing Secret: %s%s", m.mount.Path, m.path)
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
			Render(styles.IconFlag + " unmasked secrets")
	}
	return ""
}
