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
	"github.com/lrstanley/vex/internal/ui"
)

var (
	version = "master"
	cli     = &Flags{}
)

type Flags struct {
	Logging     logging.Flags `embed:""`
	EnablePprof bool          `help:"enable pprof debugging server"`
}

func main() {
	config.InitConfigPath()

	_ = kong.Parse(
		cli,
		kong.Name("vex"),
		kong.Description("Terminal UI for HashiCorp Vault"),
		kong.UsageOnError(),
		kong.Vars{
			"CONFIG_PATH": config.GetConfigPath(),
		},
	)

	closer := logging.New(version, cli.Logging)
	defer closer()

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
		os.Exit(1)
	}

	tui := tea.NewProgram(
		ui.New(client),
		tea.WithAltScreen(),
		tea.WithUniformKeyLayout(),
	)

	defer logging.RecoverPanic("main", tui.Quit)

	_, err = tui.Run()
	if err != nil {
		slog.Error("failed to run tui", "error", err)
	}
}
