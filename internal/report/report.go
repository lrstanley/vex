// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

//nolint:forbidigo
package report

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/config"
)

// Generate collects system information and outputs it to stdout.
func Generate() {
	fmt.Println("=== system info ===")
	fmt.Println("")

	fmt.Println("Application:")
	fmt.Printf("  Name:    %s\n", config.AppName)
	fmt.Printf("  Version: %s\n", config.AppVersion)
	fmt.Println("")

	fmt.Println("System:")
	fmt.Printf("  OS:         %s\n", runtime.GOOS)
	fmt.Printf("  Arch:       %s\n", runtime.GOARCH)
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	fmt.Printf("  Config:     %s\n", config.GetConfigPath())
	fmt.Println("")

	if runtime.GOOS == "linux" {
		printLinuxOSInfo()
	}
	fmt.Println("")

	printTerminalInfo()
	fmt.Println("")

	vaultInfo()
}

func printLinuxOSInfo() {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		fmt.Printf("  OS Release: unable to read /etc/os-release: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	osInfo := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "="); idx != -1 {
			key := line[:idx]
			value := strings.Trim(line[idx+1:], `"`)
			osInfo[key] = value
		}
	}

	if name, ok := osInfo["PRETTY_NAME"]; ok {
		fmt.Printf("  Distribution: %s\n", name)
	}
	if version, ok := osInfo["VERSION"]; ok {
		fmt.Printf("  Version:      %s\n", version)
	}
	if codename, ok := osInfo["VERSION_CODENAME"]; ok {
		fmt.Printf("  Codename:     %s\n", codename)
	}
}

func printTerminalInfo() {
	fmt.Println("Terminal:")

	if term := os.Getenv("TERM"); term != "" {
		fmt.Printf("  TERM:     %s\n", term)
	}
	if termProgram := os.Getenv("TERM_PROGRAM"); termProgram != "" {
		fmt.Printf("  Program:  %s\n", termProgram)
	}
	if termVersion := os.Getenv("TERM_PROGRAM_VERSION"); termVersion != "" {
		fmt.Printf("  Version:  %s\n", termVersion)
	}

	if shell := os.Getenv("SHELL"); shell != "" {
		shellName := filepath.Base(shell)
		fmt.Printf("  Shell:    %s (%s)\n", shellName, shell)
	}

	detectMultiplexer()

	if locale := os.Getenv("LANG"); locale != "" {
		fmt.Printf("  Locale:   %s\n", locale)
	}
}

func detectMultiplexer() {
	if tmux := os.Getenv("TMUX"); tmux != "" {
		fmt.Println("  Multiplexer: tmux")
		return
	}
	if screen := os.Getenv("STY"); screen != "" {
		fmt.Println("  Multiplexer: screen")
		return
	}
	fmt.Println("  Multiplexer: unknown or n/a")
}

func vaultInfo() {
	fmt.Println("Vault Health:")

	cfg := vapi.DefaultConfig()
	cfg.MaxRetries = 5
	cfg.DisableRedirects = false
	cfg.Timeout = 5 * time.Second

	if cfg.Error != nil {
		fmt.Fprintf(os.Stderr, "  failed to create vault client: %v", cfg.Error)
		return
	}

	c, err := vapi.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  failed to create vault client: %v", err)
		return
	}

	health, err := c.Sys().Health()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  failed to get health: %v", err)
	} else {
		fmt.Printf("  Version:         %s\n", health.Version)
		fmt.Printf("  Initialized:     %t\n", health.Initialized)
		fmt.Printf("  Sealed:          %v\n", health.Sealed)
		fmt.Printf("  Standby:         %v\n", health.Standby)
		fmt.Printf("  Enterprise:      %v\n", health.Enterprise)
		fmt.Printf("  Repl. DR Mode:   %s\n", health.ReplicationDRMode)
		fmt.Printf("  HA Conn Healthy: %v\n", health.HAConnectionHealthy)
		fmt.Printf("  Perf Standby:    %v\n", health.PerformanceStandby)
	}

	fmt.Println("")

	tokenSelf, err := c.Auth().Token().LookupSelf()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  failed to lookup token: %v", err)
	} else {
		path, ok := tokenSelf.Data["path"].(string)
		if ok {
			if v := strings.LastIndex(path, "/"); v != -1 {
				path = path[0:v] + "/***"
			}
			fmt.Println("  /v1/auth/token/lookup-self:")
			fmt.Printf("    Path: %s\n", path)
		}
	}
}
