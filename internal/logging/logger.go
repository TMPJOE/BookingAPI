package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new configured logger
func NewLogger(level string, format string) *Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithComponent returns a logger with the component field set
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.With(slog.String("component", component)),
	}
}

// WithRequestID returns a logger with the request ID field set
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.With(slog.String("request_id", requestID)),
	}
}
