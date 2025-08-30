// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

var truncPathTests = []struct {
	name     string
	input    string
	length   int
	expected string
}{
	{
		name:     "empty string",
		input:    "",
		length:   10,
		expected: "",
	},
	{
		name:     "single segment path",
		input:    "home",
		length:   10,
		expected: "home",
	},
	{
		name:     "path shorter than length",
		input:    "home/user",
		length:   20,
		expected: "home/user",
	},
	{
		name:     "path exactly at length",
		input:    "home/user",
		length:   9,
		expected: "home/user",
	},
	{
		name:     "path needs truncation with ellipsis",
		input:    "home/user/documents/projects/very/long/path",
		length:   25,
		expected: "home/user/…/long/path",
	},
	{
		name:     "path needs truncation from start",
		input:    "very/long/path/that/exceeds/limit",
		length:   20,
		expected: "very/…/exceeds/limit",
	},
	{
		name:     "path needs truncation from end",
		input:    "home/user/documents/projects/very/long/path",
		length:   15,
		expected: "home/…/path",
	},
	{
		name:     "path with many segments needs heavy truncation",
		input:    "a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p",
		length:   10,
		expected: "a/b/…/o/p",
	},
	{
		name:     "path with single character segments",
		input:    "a/b/c/d/e/f",
		length:   8,
		expected: "a/…/e/f",
	},
	{
		name:     "path with mixed segment lengths",
		input:    "home/user123/documents/projects/very_long_project_name",
		length:   30,
		expected: "home/…/very_long_project_name",
	},
	{
		name:     "path with leading slash",
		input:    "/home/user/documents",
		length:   15,
		expected: "/…/documents",
	},
	{
		name:     "path with trailing slash",
		input:    "home/user/documents/",
		length:   15,
		expected: "…/documents/",
	},
	{
		name:     "path with both leading and trailing slashes",
		input:    "/home/user/documents/",
		length:   15,
		expected: "/…/documents/",
	},
	{
		name:     "very short length constraint",
		input:    "home/user/documents",
		length:   5,
		expected: "…/do…",
	},
	{
		name:     "length constraint of 1",
		input:    "home/user/documents",
		length:   1,
		expected: "…",
	},
	{
		name:     "length constraint of 0",
		input:    "home/user/documents",
		length:   0,
		expected: "",
	},
	{
		name:     "path with wide characters",
		input:    "home/用户/documents/项目",
		length:   20,
		expected: "home/…/项目",
	},
	{
		name:     "path with emojis",
		input:    "home/🚀/documents/📁",
		length:   15,
		expected: "home/…/📁",
	},
}

func FuzzTruncPath(f *testing.F) {
	for _, tt := range truncPathTests {
		f.Add(tt.input, tt.length)
	}

	f.Fuzz(func(t *testing.T, input string, length int) {
		_ = TruncPath(input, length)
		// TODO: https://github.com/charmbracelet/x/issues/541
		// w := ansi.StringWidth(s)
		// if length >= 0 && w > length {
		// 	t.Errorf("TruncPath(%q, %d) = %q (len: %d), which exceeds limit %d", input, length, s, w, length)
		// }
	})
}

func TestTruncPath(t *testing.T) {
	t.Parallel()

	for _, tt := range truncPathTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := TruncPath(tt.input, tt.length)

			if result != tt.expected {
				t.Errorf("TruncPath(%q, %d) = %q (len: %d), want %q (len: %d)", tt.input, tt.length, result, ansi.StringWidth(result), tt.expected, ansi.StringWidth(tt.expected))
			}

			resultWidth := ansi.StringWidth(result)
			if resultWidth > tt.length {
				t.Errorf("TruncPath(%q, %d) returned string with width %d, which exceeds limit %d", tt.input, tt.length, resultWidth, tt.length)
			}
		})
	}
}
