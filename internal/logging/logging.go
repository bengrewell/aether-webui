package _logging

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// Options configures the global logger.
type Options struct {
	Level      slog.Level // Default: Info
	AddSource  bool       // Show file:line (tied to --debug)
	NoColor    bool       // Disable colored output
	Writer     io.Writer  // Default: os.Stderr
	TimeFormat string     // Default: time.DateTime
}

// Setup creates a tint handler and sets it as the default slog logger.
func Setup(opts Options) {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}
	tf := opts.TimeFormat
	if tf == "" {
		tf = time.DateTime
	}

	handler := tint.NewHandler(w, &tint.Options{
		Level:      opts.Level,
		AddSource:  opts.AddSource,
		NoColor:    opts.NoColor,
		TimeFormat: tf,
	})

	slog.SetDefault(slog.New(handler))
}

// With returns a child logger with the given key-value pairs attached.
func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}
