// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"sync"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
)

type PageState interface {
	// Init handles any initialisation that needs to be done for page state tracking.
	Init() tea.Cmd

	// Update handles any updates that need to be done for page state tracking, including
	// propagating messages to the active page (or inactive pages if applicable).
	Update(msg tea.Msg) tea.Cmd

	// View returns the view for the active page.
	View() string

	// All returns all pages in the page stack.
	All() []Page

	// Get returns the current active page.
	Get() Page

	// HasParent returns whether the page has a parent page.
	HasParent() bool
}

type Page interface {
	// UUID returns the UUID of the page.
	UUID() string

	// Init handles any initialisation that needs to be done for page.
	Init() tea.Cmd

	// Update handles any updates that need to be done for the page.
	Update(msg tea.Msg) tea.Cmd

	// View returns the view for the page.
	View() string

	// HasInputFocus returns whether the page captures focus for keybinds (and we
	// should send all keybinds to the page). This is stubbed to false by default.
	HasInputFocus() bool

	// GetCommands returns the commands of the page (if one is defined).
	GetCommands() []string

	// GetSupportFiltering returns whether the page supports filtering.
	GetSupportFiltering() bool

	// ShortHelp returns the short help for the page.
	ShortHelp() []key.Binding

	// FullHelp returns the full help for the page.
	FullHelp() [][]key.Binding

	// GetTitle returns the title of the page.
	GetTitle() string

	// Close is a callback which is invoked when the page is closed. Defaults
	// to a no-op but can be overridden by the page implementation.
	Close() tea.Cmd
}

type PageModel struct {
	once sync.Once
	uuid string

	Commands         []string
	SupportFiltering bool
	ShortKeyBinds    []key.Binding
	FullKeyBinds     [][]key.Binding
}

func (b *PageModel) UUID() string {
	b.once.Do(func() {
		b.uuid = UUID()
	})
	return b.uuid
}

func (b *PageModel) GetCommands() []string {
	return b.Commands
}

func (b *PageModel) GetSupportFiltering() bool {
	return b.SupportFiltering
}

func (b *PageModel) ShortHelp() []key.Binding {
	return b.ShortKeyBinds
}

func (b *PageModel) FullHelp() [][]key.Binding {
	return b.FullKeyBinds
}

// GetTitle is the title of the page. Defaults to the command of the page if defined,
// otherwise the ID. Can be overridden by the page.
func (b *PageModel) GetTitle() string {
	if len(b.Commands) > 0 {
		return b.Commands[0]
	}
	return b.UUID()
}

func (b *PageModel) Close() tea.Cmd {
	return nil
}

func (b *PageModel) HasInputFocus() bool {
	return false
}

// PageMsg is a wrapper for any message relating to page state.
type PageMsg struct {
	Msg any
}

type OpenPageMsg struct {
	Page Page
	Root bool
}

func OpenPage(p Page, isRoot bool) tea.Cmd {
	return CmdMsg(PageMsg{Msg: OpenPageMsg{Page: p, Root: isRoot}})
}

type CloseActivePageMsg struct{}

func CloseActivePage() tea.Cmd {
	return CmdMsg(PageMsg{Msg: CloseActivePageMsg{}})
}

// PageFocusedMsg is sent when the active page is focused, to the active page ONLY.
// Should always be wrapped in a PageMsg.
type PageFocusedMsg struct{}

// PageBlurredMsg is sent when the active page is blurred, to the active page ONLY.
// Should always be wrapped in a PageMsg.
type PageBlurredMsg struct{}
