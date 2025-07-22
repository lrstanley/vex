// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package dialogs

import (
	"github.com/lrstanley/vex/internal/types"
)

func (s *state) calcDialogSize(wh, ww int, size types.DialogSize) (height, width int) {
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

func (s *state) calcDialogPosition(wh, ww int, height, width int) (x, y int) {
	height += 2 + s.titleStyle.GetHeight() + s.titleStyle.GetVerticalFrameSize() // + s.dialogStyle.GetVerticalFrameSize() -- +2 for x.Borderize()
	width += 2                                                                   // s.dialogStyle.GetHorizontalFrameSize() -- +2 for x.Borderize()

	if wh == 0 || ww == 0 || height == 0 || width == 0 || height > wh || width > ww {
		return 0, 0
	}
	return (ww - width) / 2, (wh - height) / 2
}
