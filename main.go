// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/kong"
	"github.com/lrstanley/clix/v2"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/config"
	"github.com/lrstanley/vex/internal/logging"
	"github.com/lrstanley/vex/internal/report"
	"github.com/lrstanley/vex/internal/ui"
	"github.com/lrstanley/x/logging/handlers"
)

var cli = clix.New(
	clix.WithEnvFiles[Flags](),
	clix.WithVersionPlugin[Flags](),
	clix.WithMarkdownPlugin[Flags](),
	clix.WithKongOptions[Flags](
		kong.Vars{
			"CONFIG_PATH": config.GetConfigPath(),
		},
	),
	clix.WithAppInfo[Flags](clix.AppInfo{
		Name:        config.AppName,
		Version:     config.AppVersion,
		Description: "Terminal UI for HashiCorp Vault",
		Links:       clix.GithubLinks("github.com/lrstanley/vex", "master", "https://liam.sh"),
	}),
)

type Flags struct {
	Logging               logging.Flags `embed:""`
	EnablePprof           bool          `help:"enable pprof debugging server"`
	MaxConcurrentRequests int           `env:"MAX_CONCURRENT_REQUESTS" default:"10" help:"maximum number of concurrent requests to the vault server"`

	Report struct{} `cmd:"" help:"print system information for issue reporting"`
	UI     struct{} `cmd:"" default:"1" hidden:"" help:"start the terminal UI (default)"`
}

func main() {
	// Set up a defer to ensure that we can exit with a non-zero code if we need to,
	// while still allowing the defer stack to unwind.
	returnCode := 0
	defer func() { os.Exit(returnCode) }()

	config.InitConfigPath()

	switch cli.Context.Command() {
	case "report":
		report.Generate()
		return
	}

	logCloser := logging.New(config.AppVersion, cli.Flags.Logging)
	defer logCloser() //nolint:errcheck

	if cli.Flags.EnablePprof {
		go func() {
			slog.Info("pprof server starting on http://localhost:6060")
			err := http.ListenAndServe("localhost:6060", nil) //nolint:gosec
			if err != nil {
				slog.Error("failed to start pprof server", "error", err)
			}
		}()
	}

	client, err := api.NewClient(slog.Default(), cli.Flags.MaxConcurrentRequests)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create vault client: %v\n", err)
		returnCode = 1
		return
	}

	tui := tea.NewProgram(
		ui.New(client),
		tea.WithFilter(ui.DownsampleMouseEvents),
	)

	panicCloser := handlers.NewPanicCatcher(handlers.PanicPathName(config.GetConfigPath(), config.AppName))
	defer panicCloser(tui.Kill) //nolint:errcheck

	_, err = tui.Run()
	if err != nil {
		if errors.Is(err, tea.ErrProgramPanic) {
			panic(err)
		}

		slog.Error("failed to run tui", "error", err) //nolint:sloglint
		returnCode = 1
	}
}
