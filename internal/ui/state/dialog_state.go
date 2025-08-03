// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package state

import (
	"slices"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const (
	DialogWinHPadding = 2
	DialogWinVPadding = 2
)

var _ types.DialogState = &dialogState{}

type dialogState struct {
	// Core state.
	windowHeight int
	windowWidth  int
	dialogs      *types.OrderedMap[string, types.Dialog]

	// Styles.
	titleStyle lipgloss.Style
	// dialogStyle lipgloss.Style
}

func NewDialogState() types.DialogState {
	s := &dialogState{
		dialogs: types.NewOrderedMap[string, types.Dialog](),
	}
	s.initStyles()
	return s
}

func (s *dialogState) initStyles() {
	s.titleStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.DialogFg()).
		Padding(0, 1).
		Height(2)

	// s.dialogStyle = lipgloss.NewStyle().
	// 	Border(lipgloss.RoundedBorder()).
	// 	BorderForeground(styles.Theme.DialogBorderFg())
}

func (s *dialogState) Init() tea.Cmd {
	return nil
}

func (s *dialogState) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	var active, all bool

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.windowHeight = msg.Height
		s.windowWidth = msg.Width

		for _, dialog := range s.dialogs.Values() {
			cmds = append(cmds, s.sendDialogSize(dialog))
		}

		return tea.Batch(cmds...)
	case types.DialogMsg:
		switch msg := msg.Msg.(type) {
		case types.OpenDialogMsg:
			if s.Len() > 0 {
				_, current := s.dialogs.Peek()
				if current.DisablesChildren() {
					return types.FocusChange(types.FocusDialog)
				}
			}

			if current, exists := s.dialogs.Get(msg.Dialog.UUID()); exists {
				// Delete and reuse the existing dialog.
				s.dialogs.Delete(msg.Dialog.UUID())
				msg.Dialog = current
			}

			s.dialogs.Set(msg.Dialog.UUID(), msg.Dialog)

			return tea.Batch(
				msg.Dialog.Init(),
				s.sendDialogSize(msg.Dialog),
				types.FocusChange(types.FocusDialog),
			)
		case types.CloseActiveDialogMsg:
			if s.Len() == 0 {
				return nil
			}

			_, dialog := s.dialogs.Pop()
			if dialog == nil {
				return nil
			}

			if s.dialogs.Len() == 0 {
				return tea.Sequence(
					types.FocusChange(types.FocusPage),
					dialog.Close(),
				)
			}

			return dialog.Close()
		}
	case tea.KeyMsg:
		if s.Len() > 0 && !s.Get().HasInputFocus() {
			switch {
			case key.Matches(msg, types.KeyCancel):
				return types.CloseActiveDialog()
			case key.Matches(msg, types.KeyQuit):
				return types.AppQuit()
			}
		}
		active = true
	case tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
		active = true
	default:
		all = true
	}

	if all {
		for _, dialog := range s.dialogs.Values() {
			cmds = append(cmds, dialog.Update(msg))
		}
	} else if active && s.Len() > 0 {
		cmds = append(cmds, s.Get().Update(msg))
	}

	return tea.Batch(cmds...)
}

func (s *dialogState) sendDialogSize(dialog types.Dialog) tea.Cmd {
	h, w := s.calcDialogSize(
		s.windowHeight,
		s.windowWidth,
		dialog.GetSize(),
	)

	return dialog.Update(tea.WindowSizeMsg{
		Height: h,
		Width:  w,
	})
}

func (s *dialogState) Len() int {
	return s.dialogs.Len()
}

func (s *dialogState) Get(skipIDs ...string) types.Dialog {
	if s.Len() == 0 {
		return nil
	}
	if len(skipIDs) == 0 {
		_, dialog := s.dialogs.Peek()
		return dialog
	}

	dialogs := s.dialogs.Values()
	slices.Reverse(dialogs)
	for _, dialog := range dialogs {
		if !slices.Contains(skipIDs, dialog.UUID()) {
			return dialog
		}
	}
	return nil
}

func (s *dialogState) UUID() string {
	if s.Len() == 0 {
		return ""
	}
	id, _ := s.dialogs.Peek()
	return id
}

func (s *dialogState) GetLayers() []*lipgloss.Layer {
	dialogs := s.dialogs.Values()
	if len(dialogs) == 0 {
		return nil
	}

	layers := make([]*lipgloss.Layer, 0, len(dialogs))
	var view string

	for _, dialog := range dialogs {
		view = dialog.View()
		if view == "" {
			continue
		}
		dx, dy := s.calcDialogPosition(
			s.windowHeight,
			s.windowWidth,
			dialog.GetHeight(),
			dialog.GetWidth(),
		)

		layers = append(
			layers,
			lipgloss.NewLayer(
				styles.Border(
					lipgloss.JoinVertical(
						lipgloss.Top,
						s.titleStyle.Render(styles.Title(
							dialog.GetTitle(),
							dialog.GetWidth()-s.titleStyle.GetHorizontalFrameSize(),
							"/",
							styles.Theme.DialogTitleFg(),
							styles.Theme.DialogTitleFromFg(),
							styles.Theme.DialogTitleToFg(),
						)),
						view,
					),
					nil,
					dialog,
				),
			).X(dx).Y(dy),
		)
	}
	return layers
}

func (s *dialogState) calcDialogSize(wh, ww int, size types.DialogSize) (height, width int) {
	if size == "" {
		size = types.DialogSizeMedium
	}

	switch size {
	case types.DialogSizeSmall:
		height = min(wh-DialogWinVPadding, 10)
		width = min(ww-DialogWinHPadding, 50)
	case types.DialogSizeMedium:
		height = min(wh-DialogWinVPadding, 18)
		width = min(ww-DialogWinHPadding, 70)
	case types.DialogSizeLarge:
		height = min(wh-DialogWinVPadding, 25)
		width = min(ww-DialogWinHPadding, 90)
	case types.DialogSizeFull:
		height = wh - DialogWinVPadding
		width = ww - DialogWinHPadding
	case types.DialogSizeCustom:
		height = wh
		width = ww
	}

	return height - s.titleStyle.GetHeight() - s.titleStyle.GetVerticalFrameSize(), width
}

func (s *dialogState) calcDialogPosition(wh, ww int, height, width int) (x, y int) {
	height += 2 + s.titleStyle.GetHeight() + s.titleStyle.GetVerticalFrameSize() // + s.dialogStyle.GetVerticalFrameSize() -- +2 for x.Borderize()
	width += 2                                                                   // s.dialogStyle.GetHorizontalFrameSize() -- +2 for x.Borderize()

	if wh == 0 || ww == 0 || height == 0 || width == 0 || height > wh || width > ww {
		return 0, 0
	}
	return (ww - width) / 2, (wh - height) / 2
}
