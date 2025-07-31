// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package testui

import (
	"bytes"
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

const (
	DefaultTermHeight = 24
	DefaultTermWidth  = 80
)

type TestModel struct {
	*teatest.TestModel
	model   tea.ViewModel
	profile colorprofile.Profile
}

func (m *TestModel) View(t testing.TB) string {
	t.Helper()
	return m.model.View()
}

func (m *TestModel) ExpectViewSnapshot(t testing.TB) {
	t.Helper()
	RequireEqual(t, m.View(t), m.profile)
}

func (m *TestModel) WaitFor(t testing.TB, condition func(bts []byte) bool, opts ...teatest.WaitForOption) {
	t.Helper()
	teatest.WaitFor(t, m.Output(), condition, opts...)
}

func (m *TestModel) ExpectContains(t testing.TB, substr ...string) {
	t.Helper()
	teatest.WaitFor(t, m.Output(), func(bts []byte) bool {
		for _, substr := range substr {
			if !bytes.Contains(bts, []byte(substr)) {
				return false
			}
		}
		return true
	})
}

func (m *TestModel) ExpectViewDimensions(t testing.TB, width, height int) {
	t.Helper()
	m.ExpectViewHeight(t, height)
	m.ExpectViewWidth(t, width)
}

func (m *TestModel) ExpectViewHeight(t testing.TB, height int) {
	t.Helper()
	v := m.View(t)
	if lipgloss.Height(v) != height {
		t.Fatalf("expected height %d, got %d", height, lipgloss.Height(v))
	}
}

func (m *TestModel) ExpectViewWidth(t testing.TB, width int) {
	t.Helper()
	v := m.View(t)
	if lipgloss.Width(v) != width {
		t.Fatalf("expected width %d, got %d", width, lipgloss.Width(v))
	}
}

type RootModel interface {
	tea.Model
	tea.ViewModel
}

type NonRootModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View() string
}

type NonRootModelWrapper struct {
	model NonRootModel
}

var _ RootModel = (*NonRootModelWrapper)(nil)

func (m *NonRootModelWrapper) Init() tea.Cmd {
	return m.model.Init()
}

func (m *NonRootModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.model.Update(msg)
}

func (m *NonRootModelWrapper) View() string {
	return m.model.View()
}

func NewNonRootModel(t testing.TB, model NonRootModel, color bool, opts ...teatest.TestOption) *TestModel {
	t.Helper()
	return NewRootModel(t, &NonRootModelWrapper{model: model}, color, opts...)
}

func NewRootModel(t testing.TB, model RootModel, color bool, opts ...teatest.TestOption) *TestModel {
	t.Helper()

	profile := colorprofile.Ascii
	if color {
		profile = colorprofile.TrueColor
	}

	opts = append(
		[]teatest.TestOption{
			WithTermSize(DefaultTermWidth, DefaultTermHeight),
			teatest.WithProgramOptions(tea.WithColorProfile(profile)),
		},
		opts...,
	)

	return &TestModel{
		TestModel: teatest.NewTestModel(t, model, opts...),
		model:     model,
		profile:   profile,
	}
}

var WithTermSize = teatest.WithInitialTermSize

func WaitFor(t testing.TB, r io.Reader, condition func(bts []byte) bool, opts ...teatest.WaitForOption) {
	t.Helper()
	teatest.WaitFor(t, r, condition, opts...)
}

func WaitForContains(t testing.TB, r io.Reader, substr ...string) {
	t.Helper()
	teatest.WaitFor(t, r, func(bts []byte) bool {
		for _, substr := range substr {
			if !bytes.Contains(bts, []byte(substr)) {
				return false
			}
		}
		return true
	})
}

func RequireEqual[T []byte | string](tb testing.TB, out T, profile colorprofile.Profile) {
	tb.Helper()

	buf := &bytes.Buffer{}

	w := &colorprofile.Writer{
		Forward: buf,
		Profile: profile,
	}

	_, err := w.Write([]byte(out))
	if err != nil {
		tb.Fatalf("failed to write view: %v", err)
	}

	golden.RequireEqual(tb, buf.Bytes())
}
