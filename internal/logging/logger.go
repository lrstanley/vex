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

// NewPanicLogger creates a new panic logger that will write to a file in the
// log directory. The file will be named "panic-<app-name>-<timestamp>.log".
//
// The returned closer should be called to ensure that the log file is cleaned up
// if no panic was caught.
//
// The callback is called when the panic logger is closed, and can be used to
// perform any additional cleanup.
func NewPanicLogger(flags Flags, cb func()) (closer func() error) {
	dir := filepath.Dir(flags.Logging.File)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create log directory:", err)
		os.Exit(1)
	}

	fn := filepath.Join(
		dir,
		fmt.Sprintf(
			"panic-%s-%s.log",
			config.AppName,
			time.Now().Format("20060102-150405"),
		),
	)

	// SetCrashOutput doesn't support [io.Writer] interface, so we HAVE to create
	// a file. The workaround for this is a defer that will delete the file if it's
	// empty. So we will have empty files while the app is running, but it does
	// avoid a bunch of useless empty files building up.
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create panic log file:", err)
		os.Exit(1)
	}

	err = debug.SetCrashOutput(f, debug.CrashOptions{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to set crash output:", err)
		os.Exit(1)
	}

	_ = f.Close() // SetCrashOutput duplicates the file descriptor, so can safely close early.

	return func() error {
		_ = debug.SetCrashOutput(nil, debug.CrashOptions{})

		if cb != nil {
			cb()
		}

		var stat os.FileInfo
		stat, err = os.Stat(fn)
		if err != nil {
			return err
		}

		// If the file is empty, remove it.
		if stat.Size() == 0 {
			return os.Remove(fn)
		} else {
			time.Sleep(1 * time.Second)
			fmt.Fprintf(os.Stderr, "\n\npanic occurred, wrote dump to %s\n", fn)
		}
		return nil
	}
}
