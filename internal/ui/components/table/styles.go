// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package table

import (
	"charm.land/lipgloss/v2"
	"github.com/lrstanley/vex/internal/ui/styles"
)

type Styles struct {
	Base                      lipgloss.Style
	NoResults                 lipgloss.Style
	Header                    lipgloss.Style
	Cell                      lipgloss.Style
	SelectedRow               lipgloss.Style
	HighlightedRow            lipgloss.Style
	HighlightedAndSelectedRow lipgloss.Style
}

func (m *Model[T]) initStyles() {
	m.styles.Base = m.providedStyles.Base.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()),
		)

	m.styles.NoResults = m.providedStyles.NoResults.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.ErrorFg()).
				Background(styles.Theme.ErrorBg()),
		).
		Padding(0, 1).
		Align(lipgloss.Center)

	m.styles.Cell = m.providedStyles.Cell.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()),
		).
		Padding(0, 1).
		UnsetMarginTop().
		UnsetMarginBottom().
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()

	m.styles.Header = m.providedStyles.Header.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.AppFg()).
				Bold(true),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderLeft().
		UnsetBorderRight()

	m.styles.SelectedRow = m.providedStyles.SelectedRow.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.InfoFg()).
				Background(styles.Theme.InfoBg()).
				Bold(true),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()

	m.styles.HighlightedRow = m.providedStyles.HighlightedRow.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.SuccessFg()).
				Background(styles.Theme.SuccessBg()),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()

	m.styles.HighlightedAndSelectedRow = m.providedStyles.HighlightedAndSelectedRow.
		Inherit(
			lipgloss.NewStyle().
				Foreground(styles.Theme.SuccessFg()).
				Background(styles.Theme.InfoBg()).
				Bold(true),
		).
		Padding(m.styles.Cell.GetPadding()).
		Margin(m.styles.Cell.GetMargin()).
		UnsetBorderTop().
		UnsetBorderRight().
		UnsetBorderBottom().
		UnsetBorderLeft()
}

func (m *Model[T]) SetStyles(s Styles) {
	m.providedStyles = s
	m.initStyles()
}

type CellStyleFunc[T Row] func(row T, baseStyle lipgloss.Style, highlighted, selected bool) lipgloss.Style

func (m *Model[T]) defaultStyleFn(baseStyle lipgloss.Style, highlighted, selected bool) lipgloss.Style {
	switch {
	case highlighted && selected:
		return m.styles.HighlightedAndSelectedRow
	case highlighted:
		return m.styles.HighlightedRow
	case selected:
		return m.styles.SelectedRow
	default:
		return m.styles.Cell
	}
}
