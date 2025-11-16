// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Flags struct {
	Logging struct {
		Level string `enum:"debug,info,warn,error" default:"info" help:"set log level (debug|info|warn|error)"`
		File  string `default:"${CONFIG_PATH}/runtime.log" help:"set log file name"`
	} `embed:"" group:"logging" prefix:"logging." envprefix:"LOGGING_"`
}

func (f *Flags) GetLogLevel() slog.Level {
	switch f.Logging.Level {
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

// New creates a new logger with the given version and flags, writing to
// [flags.Logging.File].
//
// The returned closer should be called to ensure that the log file is closed
// and the logger is flushed.
func New(version string, flags Flags) (closer func() error) {
	dir := filepath.Dir(flags.Logging.File)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create log directory:", err)
		os.Exit(1)
	}

	rotator := &lumberjack.Logger{
		Filename:   flags.Logging.File,
		MaxSize:    20, // Max size in MB.
		MaxAge:     1,  // Days.
		MaxBackups: 2,
	}

	logAttrs := []slog.Attr{
		slog.String("version", version),
	}

	b, ok := debug.ReadBuildInfo()
	if ok {
		logAttrs = append(
			logAttrs,
			slog.String("go", b.GoVersion),
		)

		for _, s := range b.Settings {
			if s.Key == "vcs.revision" {
				logAttrs = append(logAttrs, slog.String("commit", s.Value))
			}
		}
	}

	handler := slog.NewJSONHandler(rotator, &slog.HandlerOptions{
		Level:     flags.GetLogLevel(),
		AddSource: true,
	}).WithAttrs(logAttrs)

	slog.SetDefault(slog.New(handler))

	slog.Info( //nolint:sloglint
		"application initialized",
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
		"version", version,
	)

	return rotator.Close
}
