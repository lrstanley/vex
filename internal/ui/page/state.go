// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package page

import (
	"fmt"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea/v2"
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

	pages types.AtomicSlice[types.Page]
}

func NewState(initial types.Page) types.PageState {
	t := &state{}
	t.pages.Push(initial)
	return t
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
	case types.PageMsg:
		switch msg := msg.Msg.(type) {
		case types.OpenPageMsg:
			var cmds []tea.Cmd
			if msg.Root {
				for page := range s.pages.IterValues() {
					cmds = append(cmds, page.Close())
				}
				s.pages.Set([]types.Page{msg.Page})
			} else {
				if s.isPageFocused.Load() && s.pages.Len() > 0 {
					cmds = append(cmds, s.pages.Peek().Update(types.CmdMsg(types.PageMsg{Msg: types.PageBlurredMsg{}})))
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
			)
		case types.AppFocusChangedMsg:
			all = true
			if msg.ID == types.FocusPage {
				s.isPageFocused.Store(true)
				cmds = append(cmds, s.pages.Peek().Update(types.PageMsg{Msg: types.PageFocusedMsg{}}))
			} else {
				s.isPageFocused.Store(false)
				cmds = append(cmds, s.pages.Peek().Update(types.PageMsg{Msg: types.PageBlurredMsg{}}))
			}
		}
	case types.AppFilterMsg:
		s.filter = msg.Text
		active = true
	case tea.KeyMsg, tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
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
		embeddedText[styles.BottomRightBorder] = fmt.Sprintf("filter: %s", s.filter)
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
