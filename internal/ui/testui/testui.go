// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package testui

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

const (
	DefaultTermHeight = 24
	DefaultTermWidth  = 80
)

type TestModel struct {
	*teatest.TestModel
	model   tea.Model
	profile colorprofile.Profile
}

// Messages returns the messages that have been sent to the model, which only works
// when the model is a non-root model.
func (m *TestModel) Messages(t testing.TB) []tea.Msg {
	t.Helper()
	if wrapper, ok := m.model.(*NonRootModelWrapper); ok {
		return wrapper.Messages()
	}
	return nil
}

// FilterMessages filters the messages that have been sent to the model based of the
// provided type of sameType, which only works when the model is a non-root model.
func (m *TestModel) FilterMessages(t testing.TB, sameType any) (filtered []tea.Msg) {
	t.Helper()
	if wrapper, ok := m.model.(*NonRootModelWrapper); ok {
		for _, msg := range wrapper.Messages() {
			if reflect.TypeOf(msg) == reflect.TypeOf(sameType) {
				filtered = append(filtered, msg)
			}
		}
	}
	return filtered
}

// WaitForFilterMessages waits for the model to send a message of the same type as
// sameType, which only works when the model is a non-root model. If the message is
// not found within 2 seconds, it will fail the test.
func (m *TestModel) WaitForFilterMessages(t testing.TB, sameType any) (filtered []tea.Msg) {
	t.Helper()

	timeout := time.Now().Add(2 * time.Second)

	for {
		if v := m.FilterMessages(t, sameType); len(v) > 0 {
			return v
		}
		if time.Now().After(timeout) {
			t.Fatalf("expected at least one message of type %T, got none", sameType)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// View returns the view of the model. Note that for root-based models, this might not
// use the latest version of the model, where as for non-root models, it will.
func (m *TestModel) View(t testing.TB) tea.View {
	t.Helper()
	var view tea.View
	view.AltScreen = true
	view.SetContent(m.model.View())
	return view
}

func (m *TestModel) String(t testing.TB) string {
	v := m.View(t)
	switch vv := v.Content.(type) {
	case *lipgloss.Canvas:
		return vv.Render()
	case fmt.Stringer:
		return vv.String()
	default:
		panic(fmt.Sprintf("unexpected content type in view: %T", vv))
	}
}

// ExpectViewSnapshot takes a snapshot of the view of the model. This method strips
// all ANSI escape codes.
func (m *TestModel) ExpectViewSnapshot(t testing.TB) {
	t.Helper()
	ExpectSnapshotNonANSI(t, m.String(t))
}

// ExpectViewSnapshotProfile takes a snapshot of the view of the model. This
// method uses the color profile of the model.
func (m *TestModel) ExpectViewSnapshotProfile(t testing.TB) {
	t.Helper()
	ExpectSnapshotProfile(t, m.String(t), m.profile)
}

// WaitFinished waits for the app to finish. This method only returns once the
// program has finished running or when it times out.
func (m *TestModel) WaitFinished(t testing.TB, opts ...teatest.FinalOpt) {
	t.Helper()

	if len(opts) == 0 {
		opts = []teatest.FinalOpt{teatest.WithFinalTimeout(10 * time.Second)}
	}
	m.TestModel.WaitFinished(t, opts...)
}

// WaitFor waits for a condition to be met. This method only returns once the
// condition is met or when it times out.
func (m *TestModel) WaitFor(t testing.TB, condition func(bts []byte) bool, opts ...teatest.WaitForOption) {
	t.Helper()
	teatest.WaitFor(t, m.Output(), condition, opts...)
}

// ExpectContains waits for the output to contain ALL of the given substrings.
// This method only returns once the condition is met or when it times out.
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

// ExpectNotContains waits for the output to not contain ANY of the given substrings.
func (m *TestModel) ExpectNotContains(t testing.TB, substr ...string) {
	t.Helper()
	teatest.WaitFor(t, m.Output(), func(bts []byte) bool {
		for _, substr := range substr {
			if bytes.Contains(bts, []byte(substr)) {
				return false
			}
		}
		return true
	})
}

// ExpectViewContains waits for the view to contain ALL of the given substrings.
func (m *TestModel) ExpectViewContains(t testing.TB, substr ...string) {
	t.Helper()
	view := m.String(t)
	for _, v := range substr {
		if !strings.Contains(view, v) {
			t.Fatalf("expected view to contain %q, got %q", v, view)
		}
	}
}

// ExpectViewNotContains waits for the view to not contain ANY of the given substrings.
func (m *TestModel) ExpectViewNotContains(t testing.TB, substr ...string) {
	t.Helper()
	view := m.String(t)
	for _, v := range substr {
		if strings.Contains(view, v) {
			t.Fatalf("expected view to not contain %q, got %q", v, view)
		}
	}
}

// ExpectViewDimensions waits for the view to have the given dimensions.
func (m *TestModel) ExpectViewDimensions(t testing.TB, width, height int) {
	t.Helper()
	m.ExpectViewHeight(t, height)
	m.ExpectViewWidth(t, width)
}

// ExpectViewHeight waits for the view to have the given height.
func (m *TestModel) ExpectViewHeight(t testing.TB, height int) {
	t.Helper()
	v := m.String(t)
	if lipgloss.Height(v) != height {
		t.Fatalf("expected height %d, got %d", height, lipgloss.Height(v))
	}
}

// ExpectViewWidth waits for the view to have the given width.
func (m *TestModel) ExpectViewWidth(t testing.TB, width int) {
	t.Helper()
	v := m.String(t)
	if lipgloss.Width(v) != width {
		t.Fatalf("expected width %d, got %d", width, lipgloss.Width(v))
	}
}

type RootModel interface {
	tea.Model
}

type NonRootModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View() string
}

type NonRootModelWrapper struct {
	model    NonRootModel
	messages []tea.Msg
}

var _ RootModel = (*NonRootModelWrapper)(nil)

func (m *NonRootModelWrapper) Init() tea.Cmd {
	return m.model.Init()
}

func (m *NonRootModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.messages = append(m.messages, msg)
	if len(m.messages) > 100 {
		m.messages = slices.Clip(m.messages[1:])
	}
	return m, m.model.Update(msg)
}

func (m *NonRootModelWrapper) View() tea.View {
	var view tea.View
	view.AltScreen = true
	view.SetContent(m.model.View())
	return view
}

func (m *NonRootModelWrapper) Messages() []tea.Msg {
	return m.messages
}

// NewNonRootModel creates a new test model for a non-root model (e.g. components,
// pages, dialogs, etc).
func NewNonRootModel(t testing.TB, model NonRootModel, color bool, opts ...teatest.TestOption) *TestModel {
	t.Helper()
	return NewRootModel(t, &NonRootModelWrapper{model: model}, color, opts...)
}

// NewRootModel creates a new test model for a root model (e.g. the main app).
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

// WithTermSize is a test option that sets the initial terminal size.
var WithTermSize = teatest.WithInitialTermSize

// WaitFor waits for a condition to be met. This method only returns once the
// condition is met or when it times out.
func WaitFor(t testing.TB, r io.Reader, condition func(bts []byte) bool, opts ...teatest.WaitForOption) {
	t.Helper()
	teatest.WaitFor(t, r, condition, opts...)
}

// WaitForContains waits for the output to contain ALL of the given substrings.
// This method only returns once the condition is met or when it times out.
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

// WaitNotContains waits for the output to not contain ANY of the given substrings.
// This method only returns once the condition is met or when it times out.
func WaitNotContains(t testing.TB, r io.Reader, substr ...string) {
	t.Helper()
	teatest.WaitFor(t, r, func(bts []byte) bool {
		for _, substr := range substr {
			if bytes.Contains(bts, []byte(substr)) {
				return false
			}
		}
		return true
	})
}
