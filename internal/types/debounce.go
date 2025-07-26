// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Debouncer struct {
	uuid          uuid
	LastTimestamp int64
}

func (d *Debouncer) InputsUpdated() {
	d.LastTimestamp = time.Now().UnixNano()
}

func (d *Debouncer) Is(msg DebounceMsg) bool {
	return msg.UUID == d.uuid.String() && msg.Timestamp == d.LastTimestamp
}

func (d *Debouncer) Send(dur time.Duration) tea.Cmd {
	if dur == 0 {
		dur = 150 * time.Millisecond
	}
	d.InputsUpdated()
	id := d.LastTimestamp

	return MsgAfterDuration(DebounceMsg{UUID: d.uuid.String(), Timestamp: id}, dur)
}

type DebounceMsg struct {
	UUID      string
	Timestamp int64
}
