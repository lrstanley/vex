// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/config"
	"github.com/lrstanley/vex/internal/logging"
	"github.com/lrstanley/vex/internal/report"
	"github.com/lrstanley/vex/internal/ui"
)

var cli = &Flags{}

type Flags struct {
	Logging     logging.Flags `embed:""`
	EnablePprof bool          `help:"enable pprof debugging server"`

	Report struct{} `cmd:"" help:"print system information for issue reporting"`
	UI     struct{} `cmd:"" default:"1" hidden:"" help:"start the terminal UI (default)"`
}

func main() {
	// Set up a defer to ensure that we can exit with a non-zero code if we need to,
	// while still allowing the defer stack to unwind.
	returnCode := 0
	defer func() {
		os.Exit(returnCode)
	}()

	config.InitConfigPath()

	cctx := kong.Parse(
		cli,
		kong.Name("vex"),
		kong.Description("Terminal UI for HashiCorp Vault"),
		kong.UsageOnError(),
		kong.Vars{
			"CONFIG_PATH": config.GetConfigPath(),
		},
	)

	switch cctx.Command() {
	case "report":
		report.Generate()
		return
	}

	logCloser := logging.New(config.AppVersion, cli.Logging)
	defer logCloser() //nolint:errcheck

	if cli.EnablePprof {
		go func() {
			slog.Info("pprof server starting on http://localhost:6060")
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				slog.Error("failed to start pprof server", "error", err)
			}
		}()
	}

	client, err := api.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create vault client: %v\n", err)
		returnCode = 1
		return
	}

	tui := tea.NewProgram(
		ui.New(client),
		tea.WithAltScreen(),
		tea.WithUniformKeyLayout(),
	)

	panicCloser := logging.NewPanicLogger(cli.Logging, tui.Kill)
	defer panicCloser() //nolint:errcheck

	_, err = tui.Run()
	if err != nil {
		slog.Error("failed to run tui", "error", err)
		returnCode = 1
		return
	}
}
