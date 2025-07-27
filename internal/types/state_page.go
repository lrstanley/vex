// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"time"

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

	// UUIDs returns all the UUIDs of the pages in the page stack.
	UUIDs() []string

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

	// GetRefreshInterval returns the refresh interval for the page. Refresh
	// support only configured if interval is greater than 0.
	GetRefreshInterval() time.Duration

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
	uuid uuid

	Commands         []string
	SupportFiltering bool
	RefreshInterval  time.Duration
	ShortKeyBinds    []key.Binding
	FullKeyBinds     [][]key.Binding
}

func (b *PageModel) UUID() string {
	return b.uuid.String()
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

func (b *PageModel) GetRefreshInterval() time.Duration {
	return b.RefreshInterval
}

type OpenPageMsg struct {
	Page Page
	Root bool
}

func OpenPage(p Page, isRoot bool) tea.Cmd {
	return CmdMsg(OpenPageMsg{Page: p, Root: isRoot})
}

type CloseActivePageMsg struct{}

func CloseActivePage() tea.Cmd {
	return CmdMsg(CloseActivePageMsg{})
}

// PageVisibleMsg is sent when the active page is made visible, to the active page ONLY.
// It is not send on initial page creation, only in situations like when a child page is
// closed, and the parent page is made visible again.
type PageVisibleMsg struct{}

// PageHiddenMsg is sent when the active page is made hidden, to the active page ONLY.
// It's sent when the page is no longer being actively rendered (e.g. child page is being
// displayed).
type PageHiddenMsg struct{}

// PageRefocusedMsg is sent when the active page is refocused, to the active page ONLY.
// It's not sent on initial page load, only when the page is refocused.
type PageRefocusedMsg struct{}

// PageBlurredMsg is sent when the active page is blurred, to the active page ONLY. Sent
// when things like a dialog is opened in front of the page.
type PageBlurredMsg struct{}
