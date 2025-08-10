package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
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
	// Use the global levelVar so level can be changed at runtime
	levelVar.Set(parseLevel(opts.Level))

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	hOpts := &slog.HandlerOptions{
		Level:     &levelVar,
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

var (
	defaultLogger *slog.Logger
	levelOnce     sync.Once
	initOnce      sync.Once
	lazyOnce      sync.Once
)

// levelVar is the global slog.LevelVar powering the logger's level.
var levelVar slog.LevelVar

// Init creates a logger from Options and sets it as the package/global default.
// It also calls slog.SetDefault so packages using slog.Default() get the same logger.
func Init(opts Options) *slog.Logger {
	// If already initialized, only update level and return existing logger
	if defaultLogger != nil {
		levelVar.Set(parseLevel(opts.Level))
		return defaultLogger
	}
	var l *slog.Logger
	initOnce.Do(func() {
		// Ensure levelVar has an initial value once
		levelOnce.Do(func() { levelVar.Set(parseLevel(opts.Level)) })
		l = New(opts)
		SetDefault(l)
	})
	if defaultLogger != nil {
		return defaultLogger
	}
	// Fallback in unlikely race: build and set
	if l == nil {
		l = New(opts)
	}
	SetDefault(l)
	return l
}

// SetDefault sets the provided logger as the package/global default and
// updates slog's global default as well.
func SetDefault(l *slog.Logger) {
	if l == nil {
		return
	}
	defaultLogger = l
	slog.SetDefault(l)
}

// L returns the default logger, initializing a sensible one if not set yet.
func L() *slog.Logger {
	if defaultLogger != nil {
		return defaultLogger
	}
	lazyOnce.Do(func() {
		if defaultLogger == nil {
			// Set a sane default and build logger
			levelOnce.Do(func() { levelVar.Set(parseLevel("info")) })
			defaultLogger = New(Options{Level: "info", Format: "json", AddSource: false})
			slog.SetDefault(defaultLogger)
		}
	})
	return defaultLogger
}

// SetLevel updates the global log level at runtime (debug/info/warn/error).
func SetLevel(level string) {
	levelVar.Set(parseLevel(level))
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
