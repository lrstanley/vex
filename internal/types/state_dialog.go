// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type DialogState interface {
	// Init handles any initialisation that needs to be done for the dialog state.
	// This is called when the dialog state is first created.
	Init() tea.Cmd

	// Update handles any updates that need to be done for the dialog state,
	// propagating messages to the active dialog (or inactive dialogs if applicable).
	Update(msg tea.Msg) tea.Cmd

	// Len returns the number of dialogs in the dialog state.
	Len(skipCore bool) int

	// Get returns the currently active dialog. If skipCore is true, the dialog
	// will be returned if it is not a core dialog (e.g. help, commander, etc).
	Get(skipCore bool) Dialog

	// ShortHelp returns the short help for the dialog state.
	ShortHelp() []key.Binding

	// FullHelp returns the full help for the dialog state.
	FullHelp() [][]key.Binding

	// View returns a layer for each dialog in the dialog state.
	View() []*lipgloss.Layer
}

type DialogSize string

const (
	DialogSizeSmall  DialogSize = "small"
	DialogSizeMedium DialogSize = "medium"
	DialogSizeLarge  DialogSize = "large"
	DialogSizeFull   DialogSize = "full"
	DialogSizeCustom DialogSize = "custom"
)

type Dialog interface {
	// UUID returns the UUID of the dialog.
	UUID() string

	// This handles any initialisation that needs to be done for the dialog.
	// This is called when the dialog is first created.
	Init() tea.Cmd

	// This handles any updates that need to be done for the dialog, potentially
	// propagating messages to downstream components.
	Update(msg tea.Msg) tea.Cmd

	// This returns the rendered view for the dialog.
	View() string

	// HasInputFocus returns whether the page captures focus for keybinds (and we
	// should send all keybinds to the page). This is stubbed to false by default.
	HasInputFocus() bool

	// GetSize returns the size of the dialog.
	GetSize() DialogSize

	// GetHeight returns the height of the dialog.
	GetHeight() int

	// GetWidth returns the width of the dialog.
	GetWidth() int

	// GetTitle returns the title of the dialog. Defaults to the ID of the dialog,
	// but can be overridden by the dialog implementation.
	GetTitle() string

	// DisablesChildren returns whether the dialog disables children (i.e. something
	// like a quit dialog).
	DisablesChildren() bool

	// ShortHelp returns the short help for the dialog.
	ShortHelp() []key.Binding

	// FullHelp returns the full help for the dialog.
	FullHelp() [][]key.Binding

	// IsCoreDialog returns whether the dialog is a core dialog, i.e. one that should
	// by typically skipped when checking for keybinds, some propagation, etc.
	IsCoreDialog() bool

	// Close is a callback which is invoked when the dialog is closed. Defaults
	// to a no-op but can be overridden by the dialog implementation.
	Close() tea.Cmd
}

// DialogModel is the base model for a dialog.
type DialogModel struct {
	uuid uuid

	Size            DialogSize
	Height          int
	Width           int
	DisableChildren bool
	ShortKeyBinds   []key.Binding
	FullKeyBinds    [][]key.Binding
}

func (m *DialogModel) UUID() string {
	return m.uuid.String()
}

func (m *DialogModel) GetSize() DialogSize {
	return m.Size
}

func (m *DialogModel) GetHeight() int {
	return m.Height
}

func (m *DialogModel) GetWidth() int {
	return m.Width
}

func (m *DialogModel) GetTitle() string {
	return m.UUID()
}

func (m *DialogModel) DisablesChildren() bool {
	return m.DisableChildren
}

func (m *DialogModel) ShortHelp() []key.Binding {
	return m.ShortKeyBinds
}

func (m *DialogModel) FullHelp() [][]key.Binding {
	return m.FullKeyBinds
}

func (m *DialogModel) IsCoreDialog() bool {
	return false
}

func (m *DialogModel) Close() tea.Cmd {
	return nil
}

func (m *DialogModel) HasInputFocus() bool {
	return false
}

// DialogMsg is a message that is sent to a dialog.
type DialogMsg struct {
	Msg any
}

// OpenDialogMsg is a message that is sent to open a dialog.
type OpenDialogMsg struct {
	Dialog Dialog
}

func OpenDialog(d Dialog) tea.Cmd {
	return CmdMsg(DialogMsg{Msg: OpenDialogMsg{Dialog: d}})
}

// CloseActiveDialogMsg is a message that is sent to close the active dialog.
type CloseActiveDialogMsg struct{}

func CloseActiveDialog() tea.Cmd {
	return CmdMsg(DialogMsg{Msg: CloseActiveDialogMsg{}})
}
