// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package dialogs

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const (
	DialogWinHPadding = 2
	DialogWinVPadding = 2
)

var _ types.DialogState = &state{}

type state struct {
	// Core state.
	windowHeight int
	windowWidth  int
	dialogs      *types.OrderedMap[string, types.Dialog]

	// Styles.
	titleStyle lipgloss.Style
	// dialogStyle lipgloss.Style
}

func NewState() types.DialogState {
	s := &state{
		dialogs: types.NewOrderedMap[string, types.Dialog](),
	}
	s.initStyles()
	return s
}

func (s *state) initStyles() {
	s.titleStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.DialogFg()).
		Padding(0, 1).
		Height(2)

	// s.dialogStyle = lipgloss.NewStyle().
	// 	Border(lipgloss.RoundedBorder()).
	// 	BorderForeground(styles.Theme.DialogBorderFg())
}

func (s *state) Init() tea.Cmd {
	return nil
}

func (s *state) Update(msg tea.Msg) tea.Cmd {
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
		case types.CloseDialogMsg:
			if s.Len() == 0 {
				return nil
			}

			var dialog types.Dialog

			if msg.Dialog != nil {
				dialog = msg.Dialog
				s.dialogs.Delete(dialog.UUID())
			} else {
				_, dialog = s.dialogs.Pop()
			}

			if s.dialogs.Len() == 0 {
				return tea.Sequence(
					types.FocusChange(types.FocusPage),
					dialog.Close(),
				)
			}

			return dialog.Close()
		}
	case tea.KeyMsg, tea.PasteStartMsg, tea.PasteMsg, tea.PasteEndMsg:
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

func (s *state) sendDialogSize(dialog types.Dialog) tea.Cmd {
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

func (s *state) Len() int {
	return s.dialogs.Len()
}

func (s *state) Get() types.Dialog {
	if s.Len() == 0 {
		return nil
	}
	_, dialog := s.dialogs.Peek()
	return dialog
}

func (s *state) GetWithSkip(ids ...string) types.Dialog {
	dialogs := s.dialogs.Values()
	slices.Reverse(dialogs)
	for _, dialog := range dialogs {
		if !slices.Contains(ids, dialog.UUID()) {
			return dialog
		}
	}
	return nil
}

func (s *state) UUID() string {
	if s.Len() == 0 {
		return ""
	}
	id, _ := s.dialogs.Peek()
	return id
}

func (s *state) GetLayers() (layers []*lipgloss.Layer) {
	for _, dialog := range s.dialogs.Values() {
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
						dialog.View(),
					),
					nil,
					dialog,
					nil,
				),
			).X(dx).Y(dy),
		)
	}
	return layers
}
