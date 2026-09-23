package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
	file *os.File
}

func New(filename string, level string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		return nil, fmt.Errorf("failed to create/open logging file: %w", err)
	}

	writer := io.MultiWriter(file, os.Stdout)

	handler := &slog.HandlerOptions{
		Level: levelValueToSlogLevel(level),
	}

	textHandler := slog.NewTextHandler(writer, handler)
	logger := slog.New(textHandler)

	return &Logger{
		Logger: logger,
		file:   file,
	}, nil
}

func (l *Logger) Close() {
	_ = l.file.Close()
}

func levelValueToSlogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
