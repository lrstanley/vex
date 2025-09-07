// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package config

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"slices"
)

var defaultEditors = map[string][]string{
	"windows": {
		"notepad++.exe",
		"npp.exe",
		"notepad.exe",
	},
	"default": {
		"nvim",
		"hx",
		"helix",
		"vim",
		"vi",
		"nano",
		"gedit",
		"mate",
		"kate",
	},
}

// ResolveEditor resolves the editor to use for the given platform, using $EDITOR
// when available, and using a fallback list of editors when $EDITOR is not defined.
func ResolveEditor() (path string, err error) {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return exec.LookPath(editor)
	}

	editors := slices.Clone(defaultEditors["default"])
	if _, ok := defaultEditors[runtime.GOOS]; ok {
		editors = append(editors, defaultEditors[runtime.GOOS]...)
	}

	for _, editor := range editors {
		path, err = exec.LookPath(editor)
		if err == nil {
			return path, nil
		}
	}

	return "", errors.New("$EDITOR not defined")
}
