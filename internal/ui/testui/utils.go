// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package testui

import "github.com/charmbracelet/x/ansi"

// Strip strips ANSI escape codes from the given string.
func Strip(s string) string {
	return ansi.Strip(s)
}
