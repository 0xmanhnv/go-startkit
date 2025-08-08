package logger

import (
    "io"
    "log/slog"
    "os"
    "strings"
)

// Options configures the application logger.
type Options struct {
    // Level string, one of: debug, info, warn, error
    Level string
    // Format: "json" (default) or "text"
    Format string
    // AddSource adds file:line to each record
    AddSource bool
    // Output writer; defaults to os.Stdout
    Output io.Writer
}

// New creates a *slog.Logger based on Options.
func New(opts Options) *slog.Logger {
    lvl := parseLevel(opts.Level)
    lv := new(slog.LevelVar)
    lv.Set(lvl)

    if opts.Output == nil {
        opts.Output = os.Stdout
    }

    hOpts := &slog.HandlerOptions{
        Level:     lv,
        AddSource: opts.AddSource,
    }

    var handler slog.Handler
    switch strings.ToLower(opts.Format) {
    case "text":
        handler = slog.NewTextHandler(opts.Output, hOpts)
    default:
        handler = slog.NewJSONHandler(opts.Output, hOpts)
    }
    return slog.New(handler)
}

// parseLevel parses a string to slog.Level with sensible defaults.
func parseLevel(s string) slog.Level {
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "debug":
        return slog.LevelDebug
    case "warn", "warning":
        return slog.LevelWarn
    case "error", "err":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}

