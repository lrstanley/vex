// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"context"
	"crypto/md5"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/atotto/clipboard"
	tea "charm.land/bubbletea/v2"
	"github.com/lrstanley/vex/internal/config"
)

type AppState interface {
	Page() PageState
	Dialog() DialogState
	Client() Client
}

type AppQuitMsg struct{}

// AppQuit is sent when the user wants to quit the application. Don't use [tea.Quit],
// as different state may need to be cleaned up before quitting.
func AppQuit() tea.Cmd {
	return CmdMsg(AppQuitMsg{})
}

type FocusID string

const (
	FocusPage      FocusID = "page"
	FocusDialog    FocusID = "dialog"
	FocusStatusBar FocusID = "statusbar"
)

type AppFocusChangedMsg struct {
	ID FocusID
}

func FocusChange(id FocusID) tea.Cmd {
	return CmdMsg(AppFocusChangedMsg{ID: id})
}

type AppRequestPreviousFocusMsg struct{}

func RequestPreviousFocus() tea.Cmd {
	return CmdMsg(AppRequestPreviousFocusMsg{})
}

// AppFilterMsg is sent when the user provides a filter in the status bar.
type AppFilterMsg struct {
	UUID string
	Text string
}

func AppFilter(uuid, text string) tea.Cmd {
	return CmdMsg(AppFilterMsg{UUID: uuid, Text: text})
}

type AppFilterClearedMsg struct{}

func ClearAppFilter() tea.Cmd {
	return CmdMsg(AppFilterClearedMsg{})
}

func SetClipboard(content string) tea.Cmd {
	return tea.Sequence(
		// Use OSC 52 where possible, but native clipboard for fallback (and if available
		// e.g. would need xclip or xsel on linux).
		tea.SetClipboard(content),
		func() tea.Msg {
			_ = clipboard.WriteAll(content)
			return nil
		},
		SendStatus("copied to clipboard", Info, 1*time.Second),
	)
}

// OpenTempEditor opens a temporary editor for the given path template, and
// default content.
func OpenTempEditor(uuid, pathTemplate, content string, cb func(EditorResultMsg) tea.Cmd) tea.Cmd {
	editor, err := config.ResolveEditor()
	if err != nil {
		return SendStatus(err.Error(), Error, 2*time.Second)
	}

	l := slog.With( // nolint:sloglint
		"uuid", uuid,
		"editor", editor,
	)

	sumBefore := fmt.Sprintf("%x", md5.Sum([]byte(content))) // nolint:gosec

	tmpFn, err := os.CreateTemp("", config.AppName+"-"+pathTemplate)
	if err != nil {
		return SendStatus(err.Error(), Error, 2*time.Second)
	}

	l.Info("opening editor", "path", tmpFn.Name()) // nolint:sloglint

	_, err = tmpFn.WriteString(content)
	if err != nil {
		return SendStatus(err.Error(), Error, 2*time.Second)
	}
	_ = tmpFn.Close()

	cmd := exec.CommandContext(context.Background(), editor, tmpFn.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		defer func() {
			rerr := os.Remove(tmpFn.Name())
			if rerr != nil {
				l.Error("failed to remove temp file", "error", rerr)
			} else {
				l.Info("removed temp file", "path", tmpFn.Name())
			}
		}()

		if err != nil {
			return SendStatus(err.Error(), Error, 2*time.Second)
		}

		var data []byte
		data, err = os.ReadFile(tmpFn.Name())
		if err != nil {
			return SendStatus(err.Error(), Error, 2*time.Second)
		}

		sumAfter := fmt.Sprintf("%x", md5.Sum(data)) // nolint:gosec

		msg := EditorResultMsg{
			UUID:         uuid,
			Before:       content,
			After:        string(data),
			HasChanged:   sumBefore != sumAfter,
			MD5SumBefore: sumBefore,
			MD5SumAfter:  sumAfter,
		}

		if cb == nil {
			return msg
		}
		cmd := cb(msg)
		if cmd == nil {
			return nil
		} else {
			return cmd()
		}
	})
}

// EditorResultMsg is a message to indicate the result of an external editor
// operation.
type EditorResultMsg struct {
	// UUID is used to which internal model is associated with the editor operation.
	UUID string

	// Before is the original content used for the file, if any.
	Before string

	// After is the content of the file after the editor operation.
	After string

	// HasChanged is true if the content of the file has changed.
	HasChanged bool

	// MD5SumBefore is the MD5 sum of the original content.
	MD5SumBefore string

	// MD5SumAfter is the MD5 sum of the content after the editor operation.
	MD5SumAfter string
}
