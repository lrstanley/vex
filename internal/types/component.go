// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	tea "charm.land/bubbletea/v2"
)

type Component interface {
	// GetUUID returns the UUID of the component.
	UUID() string

	// Init handles any initialisation that needs to be done for the component state.
	// This is called when the component state is first created.
	Init() tea.Cmd

	// Update handles any updates that need to be done for the component state,
	// propagating messages to the active component (or inactive components if applicable).
	Update(msg tea.Msg) tea.Cmd

	// This returns the rendered view for the component.
	View() string

	// GetHeight returns the height of the component.
	GetHeight() int

	// GetWidth returns the width of the component.
	GetWidth() int

	// GetSize returns the size of the component.
	GetSize() (x, y int)
}

type ComponentModel struct {
	uuid uuid

	Height int
	Width  int
}

func (b *ComponentModel) UUID() string {
	return b.uuid.String()
}

func (b *ComponentModel) GetHeight() int {
	return b.Height
}

func (b *ComponentModel) GetWidth() int {
	return b.Width
}

func (b *ComponentModel) GetSize() (x, y int) {
	return b.Width, b.Height
}
