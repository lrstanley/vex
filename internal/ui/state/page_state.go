// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package state

import (
	"sync/atomic"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/debouncer"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/errorview"
	"github.com/lrstanley/vex/internal/ui/components/loader"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const (
	PageHPadding = 2
	PageVPadding = 2
)

var _ types.PageState = &pageState{} // Ensure state implements types.PageState.

type pageState struct {
	// Various temporary states.
	windowHeight int
	windowWidth  int
	pages        types.AtomicSlice[types.Page]

	filter  string
	focused atomic.Bool
	loading atomic.Bool
	errored atomic.Bool

	filterStyle      lipgloss.Style
	filterIconStyle  lipgloss.Style
	separatorStyle   lipgloss.Style
	refreshStyle     lipgloss.Style
	refreshIconStyle lipgloss.Style

	// Child components.
	loader    *loader.Model
	errorview *errorview.Model
}

func NewPageState(initial types.Page) types.PageState {
	t := &pageState{
		loader:    loader.New(),
		errorview: errorview.New(),
	}
	t.pages.Push(initial)
	t.setStyles()
	return t
}

func (s *pageState) setStyles() {
	s.filterStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.PageBorderFilterFg())
	s.filterIconStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.PageBorderFilterFg()).
		PaddingRight(1).
		SetString(styles.IconFilter())
	s.separatorStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.AppFg()).
		Padding(0, 1).
		SetString(styles.IconSeparator)
	s.refreshStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.PageBorderFilterFg())
	s.refreshIconStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.PageBorderFilterFg()).
		PaddingRight(1).
		SetString(styles.IconRefresh)
}

func (s *pageState) Init() tea.Cmd {
	cmds := []tea.Cmd{
		s.loader.Init(),
		s.errorview.Init(),
	}
	for page := range s.pages.IterValues() {
		cmds = append(cmds, page.Init())
	}
	return tea.Batch(cmds...)
}

func (s *pageState) Update(msg tea.Msg) tea.Cmd { //nolint:gocognit
	var cmds []tea.Cmd

	var active, all bool

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.windowHeight = msg.Height
		s.windowWidth = msg.Width
		innerHeight := s.windowHeight - PageVPadding
		innerWidth := s.windowWidth - PageHPadding

		s.loader.SetHeight(innerHeight)
		s.loader.SetWidth(innerWidth)
		s.errorview.SetWidth(innerWidth)
		s.errorview.SetHeight(innerHeight)

		for page := range s.pages.IterValues() {
			cmds = append(cmds, page.Update(tea.WindowSizeMsg{
				Height: innerHeight,
				Width:  innerWidth,
			}))
		}

		return tea.Batch(cmds...)
	case styles.ThemeUpdatedMsg:
		s.setStyles()
		cmds = append(
			cmds,
			s.loader.Update(msg),
			s.errorview.Update(msg),
		)
		all = true
	case types.PageLoadingMsg:
		s.loading.Store(true)
		s.errored.Store(false)
		return s.loader.Active()
	case types.PageErrorsMsg:
		s.errorview.SetErrors(msg.Errors...)
		s.errored.Store(true)
		s.loading.Store(false)
		return nil
	case types.PageClearStateMsg:
		s.loading.Store(false)
		s.errored.Store(false)
		return nil
	case spinner.TickMsg:
		if s.loading.Load() && s.loader.SpinnerID() == msg.ID {
			return s.loader.Update(msg)
		}
		if s.loader.SpinnerID() != msg.ID {
			active = true
		}
	case types.OpenPageMsg:
		s.loading.Store(false)
		s.errored.Store(false)

		if msg.Root {
			for page := range s.pages.IterValues() {
				cmds = append(cmds, page.Close())
			}
			s.pages.Set([]types.Page{msg.Page})
		} else {
			if s.focused.Load() && s.pages.Len() > 0 {
				cmds = append(cmds, s.pages.Peek().Update(types.CmdMsg(types.PageHiddenMsg{})))
			}
			s.pages.Push(msg.Page)
		}

		return tea.Batch(append(
			cmds,
			msg.Page.Init(),
			msg.Page.Update(tea.WindowSizeMsg{
				Height: s.windowHeight - PageVPadding,
				Width:  s.windowWidth - PageHPadding,
			}),
			types.FocusChange(types.FocusPage),
		)...)
	case types.CloseActivePageMsg:
		if s.pages.Len() <= 1 {
			return nil
		}

		s.errored.Store(false)
		s.loading.Store(false)

		page, ok := s.pages.Pop()
		if !ok {
			return nil
		}

		return tea.Batch(
			page.Close(),
			types.FocusChange(types.FocusPage),
			s.pages.Peek().Update(types.PageVisibleMsg{}),
		)
	case types.AppFocusChangedMsg:
		if msg.ID == types.FocusPage {
			s.focused.Store(true)
			cmds = append(cmds, s.pages.Peek().Update(types.PageRefocusedMsg{}))
		} else {
			s.focused.Store(false)
			cmds = append(cmds, s.pages.Peek().Update(types.PageBlurredMsg{}))
		}
		all = true
	case types.AppFilterMsg:
		s.filter = msg.Text
		active = true
	case tea.KeyMsg:
		if !s.Get().HasInputFocus() && s.focused.Load() {
			switch {
			case key.Matches(msg, types.KeyCancel):
				switch {
				case s.Get().GetSupportFiltering() && s.filter != "":
					return types.ClearAppFilter()
				case s.HasParent():
					return types.CloseActivePage()
				}
			case key.Matches(msg, types.KeyRefresh):
				return types.RefreshData(s.Get().UUID())
			case key.Matches(msg, types.KeyQuit):
				return types.AppQuit()
			}
		}
		active = true
	case types.RefreshDataMsg:
		p := s.Get()
		if p.UUID() != msg.UUID {
			return nil
		}
		cmds = append(cmds, p.Update(msg))
		if v := p.GetRefreshInterval(); v > 0 {
			cmds = append(cmds, debouncer.Send(msg.UUID, v, types.CmdMsg(msg)))
		}
		return tea.Batch(cmds...)
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		active = true
	default:
		all = true
	}

	if all {
		for page := range s.pages.IterValues() {
			cmds = append(cmds, page.Update(msg))
		}
	} else if active {
		cmds = append(cmds, s.Get().Update(msg))
	}

	return tea.Batch(cmds...)
}

func (s *pageState) View() string {
	p := s.pages.Peek()

	embeddedText := styles.BorderFromElement(p)

	if p.GetSupportFiltering() && s.filter != "" {
		embeddedText[styles.BottomRightBorder] = s.filterIconStyle.String() + s.filterStyle.Render("filter: "+formatter.Trunc(s.filter, 20))
	}

	if p.GetRefreshInterval() > 0 {
		if embeddedText[styles.BottomRightBorder] != "" {
			embeddedText[styles.BottomRightBorder] += s.separatorStyle.String()
		}
		embeddedText[styles.BottomRightBorder] += s.refreshIconStyle.String() + s.refreshStyle.Render("refresh: "+p.GetRefreshInterval().String())
	}

	var out string

	switch {
	case s.loading.Load():
		out = s.loader.View()
	case s.errored.Load():
		out = s.errorview.View()
	default:
		out = p.View()
	}

	return styles.Border(
		out,
		styles.Theme.PageBorderFg(),
		embeddedText,
	)
}

func (s *pageState) All() (pages []types.Page) {
	return s.pages.Get()
}

func (s *pageState) UUIDs() (uuids []string) {
	for _, page := range s.pages.Get() {
		uuids = append(uuids, page.UUID())
	}
	return uuids
}

func (s *pageState) Get() types.Page {
	return s.pages.Peek()
}

func (s *pageState) HasParent() bool {
	return s.pages.Len() > 1
}

func (s *pageState) IsStateFocused() bool {
	return s.focused.Load()
}

func (s *pageState) IsFocused(uuid string) bool {
	return s.focused.Load() && s.pages.Peek().UUID() == uuid
}

func (s *pageState) ShortHelp() []key.Binding {
	var prepended []key.Binding
	page := s.Get()
	keys := page.ShortHelp()

	if !types.KeyBindingContains(keys, types.KeyCommander) {
		prepended = append(prepended, types.KeyCommander)
	}

	if page.GetSupportFiltering() && !types.KeyBindingContains(keys, types.KeyFilter) {
		prepended = append(prepended, types.KeyFilter)
	}

	if !types.KeyBindingContains(keys, types.KeyHelp) {
		prepended = append(prepended, types.KeyHelp)
	}

	return append(prepended, keys...)
}

func (s *pageState) FullHelp() [][]key.Binding {
	var prepended, appended []key.Binding
	page := s.Get()
	keys := page.FullHelp()

	if s.HasParent() && !types.KeyBindingContainsFull(keys, types.KeyCancel) {
		prepended = append(prepended, types.KeyCancel)
	}

	if page.GetRefreshInterval() > 0 && !types.KeyBindingContainsFull(keys, types.KeyRefresh) {
		prepended = append(prepended, types.KeyRefresh)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyCommander) {
		appended = append(appended, types.KeyCommander)
	}

	if page.GetSupportFiltering() && !types.KeyBindingContainsFull(keys, types.KeyFilter) {
		appended = append(appended, types.KeyFilter)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyHelp) {
		appended = append(appended, types.KeyHelp)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyQuit) {
		appended = append(appended, types.KeyQuit)
	}

	if len(keys) == 0 {
		keys = [][]key.Binding{{}}
	}

	keys[len(keys)-1] = append(keys[len(keys)-1], prepended...)
	return append(keys, appended)
}
