// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// Debounce logic (example using a 500ms debounce):
//  - event -> send DebounceMsg -> set ID to 1
//  - ~150ms passed
//  - event -> send DebounceMsg -> set ID to 2
//  - ~200ms passed
//  - event -> send DebounceMsg -> set ID to 3
//  - ~150ms passed
//  - receive DebounceMsg:1 -- Is() doesn't match current ID of 3
//  - ~200ms passed
//  - receive DebounceMsg:2 -- Is() doesn't match current ID of 3
//  - ~150ms passed
//  - receive DebounceMsg:3 -- Is() matches current ID of 3
//  - since it matches, send the origin thing you wanted to send.

type Debouncer struct {
	uuid          uuid
	LastTimestamp int64
}

func (d *Debouncer) InputsUpdated() {
	d.LastTimestamp = time.Now().UnixNano()
}

// Is checks if the debounce message came from the same debouncer, and if it matches
// the last sent debounce message. If it hasn't, that means another event was sent
// after this debounce, and we should wait longer.
func (d *Debouncer) Is(msg DebounceMsg) bool {
	return msg.UUID == d.uuid.String() && msg.Timestamp == d.LastTimestamp
}

// Send sends a debounce message after the duration has passed.
func (d *Debouncer) Send(dur time.Duration) tea.Cmd {
	if dur == 0 {
		dur = 150 * time.Millisecond
	}
	d.InputsUpdated()
	id := d.LastTimestamp

	return MsgAfterDuration(DebounceMsg{UUID: d.uuid.String(), Timestamp: id}, dur)
}

// DebounceMsg is sent from a previously triggered debounce. This should be passed
// to [Debouncer.Is] to see if it came from the same debouncer, in addition to
// checking if it matches the last sent debounce message. If it hasn't, that means
// another event was sent after this debounce, and we should wait longer.
type DebounceMsg struct {
	UUID      string
	Timestamp int64
}

// DataRefreshMsg is sent when the data for a page should be refreshed.
type DataRefreshMsg struct {
	UUID string
}

// DataRefresh is a helper for triggering a data refresh. It also helps reduce the
// chance of duplicate data refreshes, as the page state tracker uses debounce logic
// to prevent duplicate refreshes.
func DataRefresh(uuid string) tea.Cmd {
	return CmdMsg(DataRefreshMsg{UUID: uuid})
}
