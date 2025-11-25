package logger

import (
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Default returns the default logger instance.
func Default() *slog.Logger {
	return defaultLogger
}

// SetDefault sets the default logger instance.
func SetDefault(l *slog.Logger) {
	defaultLogger = l
}

// Info logs at INFO level.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs at WARN level.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs at ERROR level.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Debug logs at DEBUG level.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// With returns a new logger with the given attributes.
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}
