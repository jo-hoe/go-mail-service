package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Config provides strongly-typed options for logger construction.
type Config struct {
	Level     slog.Level
	AddSource bool
	JSON      bool
}

// New creates a new slog.Logger based on the provided Config and sets it as default.
func New(cfg Config) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: cfg.Level, AddSource: cfg.AddSource}
	if cfg.JSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

// ParseLevel parses a string into a slog.Level with sensible defaults.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		fallthrough
	default:
		return slog.LevelInfo
	}
}
