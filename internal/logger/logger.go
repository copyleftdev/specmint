package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// New creates a new structured logger with sensible defaults
func New() zerolog.Logger {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"

	// Use console writer for development
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
		NoColor:    false,
	}

	return zerolog.New(output).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}

// WithLevel creates a logger with the specified level
func WithLevel(level string) zerolog.Logger {
	logger := New()
	
	switch strings.ToLower(level) {
	case "debug":
		return logger.Level(zerolog.DebugLevel)
	case "info":
		return logger.Level(zerolog.InfoLevel)
	case "warn", "warning":
		return logger.Level(zerolog.WarnLevel)
	case "error":
		return logger.Level(zerolog.ErrorLevel)
	case "fatal":
		return logger.Level(zerolog.FatalLevel)
	default:
		return logger.Level(zerolog.InfoLevel)
	}
}

// WithJSON creates a JSON logger for production use
func WithJSON() zerolog.Logger {
	return zerolog.New(os.Stderr).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}

// WithFile creates a logger that writes to a file
func WithFile(filename string) (zerolog.Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return zerolog.Logger{}, err
	}

	return zerolog.New(file).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger(), nil
}

// CorrelationID adds a correlation ID to the logger context
func CorrelationID(logger zerolog.Logger, id string) zerolog.Logger {
	return logger.With().Str("correlation_id", id).Logger()
}

// Component adds a component name to the logger context
func Component(logger zerolog.Logger, component string) zerolog.Logger {
	return logger.With().Str("component", component).Logger()
}

// Operation adds an operation name to the logger context
func Operation(logger zerolog.Logger, operation string) zerolog.Logger {
	return logger.With().Str("operation", operation).Logger()
}
