// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package state

import (
	"image"
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/components/shorthelp"
	"github.com/lrstanley/vex/internal/ui/styles"
	"github.com/lrstanley/x/charm/formatter"
)

const (
	DialogHPadding      = 2 // Borders, padding, etc of the rendered dialog itself.
	DialogVPadding      = 2 // Borders, padding, etc of the rendered dialog itself.
	DialogWindowPadding = 3 // Padding to ensure dialog doesn't touch the window edges.
)

var _ types.DialogState = &dialogState{}

type dialogState struct {
	// Core state.
	windowHeight int
	windowWidth  int
	dialogs      *types.OrderedMap[string, types.Dialog]

	// Styles.
	titleStyle lipgloss.Style

	// Child components.
	shorthelp *shorthelp.Model
}

func NewDialogState() types.DialogState {
	s := &dialogState{
		dialogs:   types.NewOrderedMap[string, types.Dialog](),
		shorthelp: shorthelp.New(),
	}
	s.initStyles()
	return s
}

func (s *dialogState) initStyles() {
	s.titleStyle = lipgloss.NewStyle().
		Foreground(styles.Theme.DialogFg()).
		Padding(0, 1, 1, 1).
		Height(2)

	helpStyles := shorthelp.Styles{}
	helpStyles.Base = helpStyles.Base.
		Foreground(styles.Theme.AppFg())
	helpStyles.Key = helpStyles.Key.
		Foreground(styles.Theme.ShortHelpKeyFg())
	helpStyles.Desc = helpStyles.Desc.
		Foreground(styles.Theme.AppFg()).
		Faint(true)
	helpStyles.Separator = helpStyles.Separator.
		Foreground(styles.Theme.AppFg()).
		Faint(true)
	s.shorthelp.SetStyles(helpStyles)
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

		return tea.Sequence(cmds...)
	case types.DialogMsg:
		switch msg := msg.Msg.(type) {
		case types.OpenDialogMsg:
			if s.dialogs.Len() > 0 {
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
			if s.dialogs.Len() == 0 {
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
		if s.dialogs.Len() > 0 && !s.Get(false).HasInputFocus() {
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
	case types.AppFocusChangedMsg:
		dialog := s.Get(true)
		if msg.ID != types.FocusDialog || dialog == nil {
			s.shorthelp.SetKeyBinds()
			return nil
		}
		s.shorthelp.SetKeyBinds(s.ShortHelp()...)
		all = true
	case styles.ThemeUpdatedMsg:
		s.initStyles()
		cmds = append(cmds, s.shorthelp.Update(msg))
		all = true
	default:
		all = true
	}

	if all {
		for _, dialog := range s.dialogs.Values() {
			cmds = append(cmds, dialog.Update(msg))
		}
	} else if active && s.dialogs.Len() > 0 {
		cmds = append(cmds, s.Get(false).Update(msg))
	}

	return tea.Batch(cmds...)
}

func (s *dialogState) sendDialogSize(dialog types.Dialog) tea.Cmd {
	h, w := s.suggestedDialogSize(
		s.windowHeight,
		s.windowWidth,
		dialog.GetSize(),
	)

	return dialog.Update(tea.WindowSizeMsg{
		Height: h,
		Width:  w,
	})
}

func (s *dialogState) Len(skipCore bool) int {
	if !skipCore {
		return s.dialogs.Len()
	}
	var out int
	for _, dialog := range s.dialogs.Values() {
		if !dialog.IsCoreDialog() {
			out++
		}
	}
	return out
}

func (s *dialogState) Get(skipCore bool) types.Dialog {
	if s.Len(skipCore) == 0 {
		return nil
	}
	if !skipCore {
		_, dialog := s.dialogs.Peek()
		return dialog
	}

	dialogs := s.dialogs.Values()
	slices.Reverse(dialogs)
	for _, dialog := range dialogs {
		if !dialog.IsCoreDialog() {
			return dialog
		}
	}
	return nil
}

func (s *dialogState) UUID() string {
	if s.dialogs.Len() == 0 {
		return ""
	}
	id, _ := s.dialogs.Peek()
	return id
}

func (s *dialogState) ShortHelp() []key.Binding {
	dialog := s.Get(true)
	if dialog == nil {
		return nil
	}

	var prepended []key.Binding
	keys := dialog.ShortHelp()

	if !types.KeyBindingContains(keys, types.KeyHelp) {
		prepended = append(prepended, types.KeyHelp)
	}

	if !types.KeyBindingContains(keys, types.KeyCancel) {
		keys = append(keys, types.KeyCancel)
	}

	return append(prepended, keys...)
}

func (s *dialogState) FullHelp() [][]key.Binding {
	dialog := s.Get(true)
	if dialog == nil {
		return nil
	}

	var prepended, appended []key.Binding
	keys := dialog.FullHelp()

	if !types.KeyBindingContainsFull(keys, types.KeyCancel) {
		prepended = append(prepended, types.KeyCancel)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyHelp) {
		appended = append(appended, types.KeyHelp)
	}

	if !types.KeyBindingContainsFull(keys, types.KeyQuit) {
		appended = append(appended, types.KeyQuit)
	}

	if len(keys) == 0 {
		keys = [][]key.Binding{{}}
	}

	keys[len(keys)-1] = append(keys[len(keys)-1], prepended...)
	return append(keys, appended)
}

func (s *dialogState) View() (layers []*lipgloss.Layer) {
	dialogs := s.dialogs.Values()
	if len(dialogs) == 0 {
		return nil
	}

	var view string
	var maxTitleWidth int
	var bounds image.Rectangle
	var embeddedText map[styles.BorderPosition]string

	for i, dialog := range dialogs {
		view = dialog.View()
		if view == "" {
			panic("dialog view is empty")
		}

		bounds = s.calcDialogPosition(
			s.windowHeight,
			s.windowWidth,
			dialog.GetHeight(),
			dialog.GetWidth(),
		)

		embeddedText = styles.BorderFromElement(dialog)

		if s.shorthelp.NumKeyBinds() > 0 && !dialog.IsCoreDialog() {
			embeddedText[styles.BottomMiddleBorder] = s.shorthelp.View()
		}

		// TODO: replace dialog.GetWidth() with bounds.Dx()
		maxTitleWidth = max(0, dialog.GetWidth()-s.titleStyle.GetHorizontalFrameSize())

		layers = append(layers, lipgloss.NewLayer(
			styles.Border(
				lipgloss.JoinVertical(
					lipgloss.Top,
					s.titleStyle.Render(styles.Title(
						// Give a little extra padding.
						formatter.TruncMaybePath(dialog.GetTitle(), maxTitleWidth-2),
						maxTitleWidth,
						styles.IconTitleGradientDivider,
						styles.Theme.TitleFg(),
						styles.Theme.TitleFromFg(),
						styles.Theme.TitleToFg(),
					)),
					view,
				),
				nil,
				embeddedText,
			),
		).Z(i+1).X(bounds.Min.X).Y(bounds.Min.Y).ID(dialog.UUID()))
	}
	return layers
}

// suggestedDialogSize returns a suggested size for a dialog based on the window size and the dialog
// size preset.
func (s *dialogState) suggestedDialogSize(wh, ww int, size types.DialogSize) (height, width int) {
	if size == "" {
		size = types.DialogSizeMedium
	}

	switch size {
	case types.DialogSizeSmall:
		height = 10
		width = 50
	case types.DialogSizeMedium:
		height = 18
		width = 70
	case types.DialogSizeLarge:
		height = 25
		width = 90
	case types.DialogSizeFull, types.DialogSizeCustom:
		height = wh
		width = ww
	}

	ch := min(height-s.titleStyle.GetHeight(), wh-s.titleStyle.GetHeight()-(DialogWindowPadding*2)-DialogVPadding)
	cw := min(width, ww-(DialogWindowPadding*2)-DialogHPadding)

	if ch < 1 || cw < 10 {
		return 0, 0
	}

	return ch, cw
}

func (s *dialogState) calcDialogPosition(wh, ww, height, width int) image.Rectangle {
	height += 2 + s.titleStyle.GetHeight() // -- +2 for x.Borderize()
	width += 2                             // -- +2 for x.Borderize()

	if wh == 0 || ww == 0 || height == 0 || width == 0 {
		return image.Rectangle{}
	}

	x := max((ww-width)/2, DialogWindowPadding)
	y := max((wh-height)/2, DialogWindowPadding)

	return image.Rectangle{
		Min: image.Pt(x, y),
		Max: image.Pt(x+width, y+height),
	}
}
