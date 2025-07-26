// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package pages

import (
	"sync/atomic"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const (
	PageHPadding = 2
	PageVPadding = 2
)

var _ types.PageState = &state{} // Ensure state implements types.PageState.

type state struct {
	// Various temporary states.
	isPageFocused atomic.Bool
	windowHeight  int
	windowWidth   int
	filter        string
	pages         types.AtomicSlice[types.Page]

	filterStyle     lipgloss.Style
	filterIconStyle lipgloss.Style
}

func NewState(initial types.Page) types.PageState {
	t := &state{}
	t.pages.Push(initial)
	t.setStyles()
	return t
}

func (s *state) setStyles() {
	s.filterStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.PageBorderFilterFg()).
		PaddingLeft(1)
	s.filterIconStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.PageBorderFilterFg())
}

func (s *state) Init() tea.Cmd {
	var cmds []tea.Cmd
	for page := range s.pages.IterValues() {
		cmds = append(cmds, page.Init())
	}
	return tea.Batch(cmds...)
}

func (s *state) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	var active, all bool

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.windowHeight = msg.Height
		s.windowWidth = msg.Width

		for page := range s.pages.IterValues() {
			cmds = append(cmds, page.Update(tea.WindowSizeMsg{
				Height: s.windowHeight - PageVPadding,
				Width:  s.windowWidth - PageHPadding,
			}))
		}

		return tea.Batch(cmds...)
	case styles.ThemeUpdatedMsg:
		s.setStyles()
		all = true
	case types.OpenPageMsg:
		var cmds []tea.Cmd
		if msg.Root {
			for page := range s.pages.IterValues() {
				cmds = append(cmds, page.Close())
			}
			s.pages.Set([]types.Page{msg.Page})
		} else {
			if s.isPageFocused.Load() && s.pages.Len() > 0 {
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
			s.isPageFocused.Store(true)
			cmds = append(cmds, s.pages.Peek().Update(types.PageRefocusedMsg{}))
		} else {
			s.isPageFocused.Store(false)
			cmds = append(cmds, s.pages.Peek().Update(types.PageBlurredMsg{}))
		}
		all = true
	case types.AppFilterMsg:
		s.filter = msg.Text
		active = true
	case tea.KeyMsg:
		if !s.Get().HasInputFocus() && s.isPageFocused.Load() {
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
				return tea.Quit
			}
		}
		active = true
	case types.DebounceMsg:
		p := s.Get()
		if p.GetRefreshDebouncer().Is(msg) {
			return types.RefreshData(p.UUID())
		}
		active = true
	case types.RefreshDataMsg:
		p := s.Get()
		if p.UUID() != msg.UUID {
			return nil
		}
		cmds = append(cmds, p.Update(msg))
		if v := p.GetRefreshInterval(); v > 0 {
			cmds = append(cmds, p.GetRefreshDebouncer().Send(v))
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

func (s *state) View() string {
	p := s.pages.Peek()

	embeddedText := make(map[styles.BorderPosition]string)
	if p.GetSupportFiltering() && s.filter != "" {
		embeddedText[styles.BottomRightBorder] = s.filterIconStyle.Render(styles.IconFilter) + s.filterIconStyle.Render("filter: "+s.filter)
	}

	return styles.Border(
		p.View(),
		styles.Theme.PageBorderFg(),
		p,
		embeddedText,
	)
}

func (s *state) All() (pages []types.Page) {
	return s.pages.Get()
}

func (s *state) UUIDs() (uuids []string) {
	for _, page := range s.pages.Get() {
		uuids = append(uuids, page.UUID())
	}
	return uuids
}

func (s *state) Get() types.Page {
	return s.pages.Peek()
}

func (s *state) HasParent() bool {
	return s.pages.Len() > 1
}

func (s *state) IsStateFocused() bool {
	return s.isPageFocused.Load()
}

func (s *state) IsFocused(uuid string) bool {
	return s.isPageFocused.Load() && s.pages.Peek().UUID() == uuid
}
