package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":     slog.LevelDebug,
		"info":      slog.LevelInfo,
		" warn ":    slog.LevelWarn,
		"error":     slog.LevelError,
		"":          slog.LevelInfo,
		"   INFO  ": slog.LevelInfo,
		"invalid":   slog.LevelInfo,
		"warning":   slog.LevelInfo,
	}

	for input, expected := range cases {
		got := ParseLevel(input)
		if got != expected {
			t.Errorf("ParseLevel(%q) = %v, want %v", input, got, expected)
		}
	}
}

func TestNewJSONLoggerOutputsJSON(t *testing.T) {
	out := captureStdout(func() {
		_ = New(Config{
			Level:     slog.LevelInfo,
			AddSource: false,
			JSON:      true,
		})
		// Ensure default logger is set and logs are emitted
		slog.Info("json test message", "k", "v")
	})

	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, "{") {
		t.Fatalf("expected JSON output to start with '{', got: %q", firstLine(trimmed))
	}
	if !strings.Contains(out, `"msg":"json test message"`) {
		t.Fatalf("expected JSON output to contain message field, got: %q", firstLine(out))
	}
	// JSON handler uses upper-case level names by default
	if !strings.Contains(out, `"level":"INFO"`) {
		t.Fatalf("expected JSON output to contain level field, got: %q", firstLine(out))
	}
	if !strings.Contains(out, `"k":"v"`) {
		t.Fatalf("expected JSON output to contain structured field k=v, got: %q", firstLine(out))
	}
}

func TestNewTextLoggerOutputsText(t *testing.T) {
	out := captureStdout(func() {
		_ = New(Config{
			Level:     slog.LevelInfo,
			AddSource: false,
			JSON:      false,
		})
		slog.Info("text test message", "k", "v")
	})

	trimmed := strings.TrimSpace(out)
	if strings.HasPrefix(trimmed, "{") {
		t.Fatalf("expected text output, got JSON-like output: %q", firstLine(trimmed))
	}
	// Text handler typically prints "level=INFO" and "msg=..."
	if !strings.Contains(out, "level=INFO") {
		t.Fatalf("expected text output to contain level=INFO, got: %q", firstLine(out))
	}
	if !(strings.Contains(out, `msg="text test message"`) || strings.Contains(out, "msg=text test message")) {
		t.Fatalf("expected text output to contain msg with text test message, got: %q", firstLine(out))
	}
	if !strings.Contains(out, "k=v") {
		t.Fatalf("expected text output to contain k=v, got: %q", firstLine(out))
	}
}

func TestLogLevelFiltering(t *testing.T) {
	// With WARN level, INFO should be filtered, WARN should pass
	out := captureStdout(func() {
		_ = New(Config{
			Level:     slog.LevelWarn,
			AddSource: false,
			JSON:      true,
		})
		slog.Info("should not appear")
		slog.Warn("should appear")
	})

	if strings.Contains(out, "should not appear") {
		t.Fatalf("INFO message should have been filtered at WARN level; output: %q", out)
	}
	if !strings.Contains(out, "should appear") {
		t.Fatalf("expected WARN message to appear; output: %q", out)
	}
	// Sanity check level in JSON
	if !strings.Contains(out, `"level":"WARN"`) {
		t.Fatalf("expected JSON output to contain level=WARN; output: %q", firstLine(out))
	}
}

func captureStdout(fn func()) string {
	// Save original stdout
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = r.Close()
	}()

	os.Stdout = w
	defer func() {
		// Restore stdout even if fn panics
		os.Stdout = orig
	}()

	fn()

	_ = w.Close()
	b, _ := io.ReadAll(r)
	return string(b)
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
