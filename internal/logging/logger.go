// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/lrstanley/vex/internal/config"
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

type logger struct {
	closer func() error
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

	slog.Info(
		"application initialized",
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
		"version", version,
	)

	return rotator.Close
}

var reCleanSrc = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// RecoverPanic recovers from panics, and logs the panic to a log file in the current
// working directory. If it's unable to write to the log file, it will log the panic
// to the standard slog location (config path).
func RecoverPanic(src string, callbacks ...func()) {
	src = reCleanSrc.ReplaceAllString(src, "")

	if r := recover(); r != nil {
		ts := time.Now()
		fn := fmt.Sprintf(
			"panic-%s-%s-%s.log",
			config.AppName,
			src,
			ts.Format("20060102-150405"),
		)

		file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			slog.Error(
				"panic occurred",
				"error", r,
				"src", src,
				"stack", string(debug.Stack()),
			)
			return
		}

		fmt.Fprintf(file, "panic @ %s: %v\n", src, r)
		fmt.Fprintf(file, "time: %s\n\n", ts.Format(time.RFC3339))
		fmt.Fprintf(file, "stack:\n%s\n", string(debug.Stack()))
		file.Close()

		for _, cb := range callbacks {
			cb()
		}

		slog.Error("application exiting due to panic", "stack-trace-path", fn)
		os.Exit(1)
	}
}
