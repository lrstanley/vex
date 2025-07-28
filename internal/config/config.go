// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName    = "vex"
	AppVersion = "0.1.0"
)

func AppTitle(subtitle string) string {
	if subtitle == "" {
		return fmt.Sprintf("%s: %s", AppName, AppVersion)
	}
	return fmt.Sprintf("%s: %s", AppName, subtitle)
}

// InitConfigPath ensures that the config path exists.
func InitConfigPath() {
	if err := os.MkdirAll(GetConfigPath(), 0o700); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create config directory:", err)
		os.Exit(1)
	}
}

// GetConfigPath returns the path to the config folder where we will store any
// app state, configurations, db, etc.
//
//   - $XDG_CONFIG_HOME/<config-folder-name> (if XDG_CONFIG_HOME is set)
//   - windows: %LOCALAPPDATA%/<config-folder-name>
//   - everything else: $HOME/.config/<config-folder-name>
func GetConfigPath() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir != "" {
		return filepath.Join(dir, AppName)
	}

	if runtime.GOOS == "windows" {
		dir = os.Getenv("LOCALAPPDATA")
		if dir != "" {
			return filepath.Join(dir, AppName)
		}
		if up := os.Getenv("USERPROFILE"); up != "" {
			return filepath.Join(up, "AppData", "Local")
		}
	}

	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", AppName)
	}

	fmt.Fprintln(os.Stderr, "failed to determine config path (no $XDG_CONFIG_HOME, $LOCALAPPDATA, $USERPROFILE, or $HOME set)")
	os.Exit(1)
	return ""
}
