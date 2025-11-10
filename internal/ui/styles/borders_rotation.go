// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package styles

import (
	"slices"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

func rotate[T any, I int | int64](arr []T, k I) {
	n := len(arr)
	k %= I(n)
	if k < 0 {
		k += I(n)
	}
	slices.Reverse(arr[:k])
	slices.Reverse(arr[k:])
	slices.Reverse(arr)
}

var (
	borderRotation    = atomic.Int64{}
	borderRotationFPS = 25
)

type BorderRotationTickMsg struct {
	Current int64
}

func BorderRotationTick(msg tea.Msg) tea.Cmd {
	// This is purely exploratory. Current design is terrible and uses a lot of CPU,
	// so may explore with a more optimized approach in the future.
	//
	//	rotate(gradient, borderRotation.Load())
	//
	v, ok := msg.(BorderRotationTickMsg)
	if msg != nil && !ok {
		return nil
	}

	if msg != nil && v.Current != borderRotation.Load() {
		return nil
	}

	return tea.Tick(time.Second/time.Duration(borderRotationFPS), func(_ time.Time) tea.Msg {
		return BorderRotationTickMsg{Current: borderRotation.Add(-3)}
	})
}
