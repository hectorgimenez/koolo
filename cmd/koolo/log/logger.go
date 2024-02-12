package log

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

var logFileHandler *os.File

func NewLogger(debug bool, logDir string) (*slog.Logger, error) {
	if logDir == "" {
		logDir = "logs"
	}

	if _, err := os.Stat(logDir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating log directory: %w", err)
		}
	}

	fileName := "koolo-log-" + time.Now().Format("2006-01-02-15-04-05") + ".txt"
	logFileHandler, err := os.Create(logDir + "/" + fileName)
	if err != nil {
		return nil, err
	}

	level := slog.LevelDebug
	if !debug {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(io.MultiWriter(os.Stdout, logFileHandler), opts)

	return slog.New(handler), nil
}

func FlushLog() error {
	if logFileHandler != nil {
		return logFileHandler.Close()
	}

	return nil
}
