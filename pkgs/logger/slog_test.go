package logger

import (
	"context"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/require"
)

func TestParseSlogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{name: "debug", input: "debug", expected: slog.LevelDebug},
		{name: "warn", input: "warn", expected: slog.LevelWarn},
		{name: "warning", input: "warning", expected: slog.LevelWarn},
		{name: "error", input: "error", expected: slog.LevelError},
		{name: "info", input: "info", expected: slog.LevelInfo},
		{name: "default", input: "unknown", expected: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, parseSlogLevel(tt.input))
		})
	}
}

func TestNewJsonSlogHandler(t *testing.T) {
	handler := NewJsonSlogHandler(nil, false)
	require.NotNil(t, handler)
	require.True(t, handler.Enabled(context.Background(), slog.LevelInfo))

	handler = NewJsonSlogHandler(&SlogConfig{Level: "error"}, false)
	require.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	require.True(t, handler.Enabled(context.Background(), slog.LevelError))
}
