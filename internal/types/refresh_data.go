// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import tea "github.com/charmbracelet/bubbletea/v2"

// RefreshDataMsg is sent when the data for a page should be refreshed.
type RefreshDataMsg struct {
	UUID string
}

// RefreshData is a helper for triggering a data refresh. It also helps reduce the
// chance of duplicate data refreshes, as the page state tracker uses debounce logic
// to prevent duplicate refreshes.
func RefreshData(uuid string) tea.Cmd {
	return CmdMsg(RefreshDataMsg{UUID: uuid})
}
