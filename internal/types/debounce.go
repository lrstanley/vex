// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Debouncer struct {
	Duration time.Duration
	ID       int64
}

func (d *Debouncer) InputsUpdated() {
	d.ID = time.Now().UnixNano()
}

func (d *Debouncer) Is(msg DebounceMsg) bool {
	return msg.ID == d.ID
}

func (d *Debouncer) Send() tea.Cmd {
	if d.Duration == 0 {
		d.Duration = 150 * time.Millisecond
	}
	d.InputsUpdated()
	id := d.ID

	return MsgAfterDuration(DebounceMsg{ID: id}, d.Duration)
}

type DebounceMsg struct {
	ID int64
}
