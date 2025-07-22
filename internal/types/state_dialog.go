// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type DialogState interface {
	// Init handles any initialisation that needs to be done for the dialog state.
	// This is called when the dialog state is first created.
	Init() tea.Cmd

	// Update handles any updates that need to be done for the dialog state,
	// propagating messages to the active dialog (or inactive dialogs if applicable).
	Update(msg tea.Msg) tea.Cmd

	// Len returns the number of dialogs in the dialog state.
	Len() int

	// Get returns the currently active dialog.
	Get() Dialog

	// GetWithSkip returns the currently active dialog, skipping the given dialog ID
	// if it's at the top of the stack.
	GetWithSkip(uuid string) Dialog

	// GetLayers returns the layers for the dialog state.
	GetLayers() []*lipgloss.Layer
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

// CloseDialogMsg is a message that is sent to close a dialog.
type CloseDialogMsg struct {
	Dialog Dialog
}

func CloseDialog(d Dialog) tea.Cmd {
	return CmdMsg(DialogMsg{Msg: CloseDialogMsg{Dialog: d}})
}
