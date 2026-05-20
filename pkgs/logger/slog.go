package logger

import (
	"log/slog"
	"os"
	"strings"
)

type SlogConfig struct {
	Level       string
	ServiceName string
}

func NewJsonSlogHandler(config *SlogConfig, sourceInfo bool) *slog.JSONHandler {
	if config == nil {
		config = &SlogConfig{
			Level:       "info",
			ServiceName: "service",
		}
	}
	if config.Level == "" {
		config.Level = "info"
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     parseSlogLevel(config.Level),
		AddSource: sourceInfo,
	})

	if config.ServiceName == "" {
		config.ServiceName = "backend-go"
	}

	return handler.WithAttrs([]slog.Attr{
		slog.String("service_name", config.ServiceName),
	}).(*slog.JSONHandler)
}

func parseSlogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info":
		fallthrough
	default:
		return slog.LevelInfo
	}
}
