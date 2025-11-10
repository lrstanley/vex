// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package debouncer

import (
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
)

const (
	JanitorFrequency = 30 * time.Second
	PruneAfter       = 5 * time.Minute
	fallbackDuration = 5 * time.Second
)

// InvokeMsg is sent when a debounce should be invoked. It will use the uuid and timestamp
// to see if it matches what is in the last-seen [DebounceMsg]. If it does, the previously
// stored [tea.Cmd] will be invoked.
type InvokeMsg struct {
	UUID      string
	Timestamp int64
}

// DebounceMsg is sent at the start of a debounce, and it used to track the debounce
// event.
type DebounceMsg struct {
	UUID      string
	Duration  time.Duration
	Cmd       tea.Cmd
	Timestamp int64
	Reset     bool
}

// Send is a helper for sending a debounce message. It will return a [DebounceMsg]
// which will be used to track the debounce event, and if no other [DebounceMsg] is
// sent to override the previous debounce events, the cmd will be invoked.
func Send(uuid string, dur time.Duration, cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return DebounceMsg{
			UUID:      uuid,
			Timestamp: time.Now().UnixNano(),
			Duration:  dur,
			Cmd:       cmd,
		}
	}
}

// ResetTimer will update the state for the uuid, which is helpful when you want
// to immediately invoke something, but still let the debouncer know it should
// wait at least the next provided duration.
func ResetTimer(uuid string) tea.Cmd {
	return func() tea.Msg {
		return DebounceMsg{
			UUID:      uuid,
			Timestamp: time.Now().UnixNano(),
			Reset:     true,
		}
	}
}

type Service struct {
	mu    sync.RWMutex
	state map[string]DebounceMsg
}

// New creates a new debouncer service, spinning up a janitor goroutine to clean up
// any expired debounce events.
func New() *Service {
	d := &Service{
		state: make(map[string]DebounceMsg),
	}
	go d.janitor()
	return d
}

func (d *Service) janitor() {
	for {
		time.Sleep(JanitorFrequency)
		d.mu.Lock()
		for k, v := range d.state {
			if time.Since(time.Unix(0, v.Timestamp)) > PruneAfter {
				delete(d.state, k)
			}
		}
		d.mu.Unlock()
	}
}

func (d *Service) Init() tea.Cmd {
	return nil
}

func (d *Service) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case DebounceMsg:
		if msg.Duration <= 0 && !msg.Reset {
			msg.Duration = fallbackDuration
		}

		d.mu.Lock()
		d.state[msg.UUID] = msg
		d.mu.Unlock()

		if msg.Reset {
			return nil
		}

		return tea.Tick(msg.Duration, func(_ time.Time) tea.Msg {
			return InvokeMsg{
				UUID:      msg.UUID,
				Timestamp: msg.Timestamp,
			}
		})
	case InvokeMsg:
		d.mu.RLock()
		v, ok := d.state[msg.UUID]
		d.mu.RUnlock()

		if !ok || v.Timestamp != msg.Timestamp {
			return nil
		}

		d.mu.Lock()
		delete(d.state, msg.UUID)
		d.mu.Unlock()
		return v.Cmd
	}
	return nil
}
