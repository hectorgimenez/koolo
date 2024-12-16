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

func NewLogger(debug bool, logDir, supervisor string) (*slog.Logger, error) {
	if logDir == "" {
		logDir = "logs"
	}

	if _, err := os.Stat(logDir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating log directory: %w", err)
		}
	}

	fileName := "Koolo-log-" + time.Now().Format("2006-01-02-15-04-05") + ".log"
	if supervisor != "" {
		fileName = fmt.Sprintf("Supervisor-log-%s-%s.log", supervisor, time.Now().Format("2006-01-02-15-04-05"))
	}

	lfh, err := os.Create(logDir + "/" + fileName)
	if err != nil {
		return nil, err
	}
	logFileHandler = lfh

	level := slog.LevelDebug
	if !debug {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.TimeKey {
				return a
			}

			t := a.Value.Time()
			a.Value = slog.StringValue(t.Format(time.TimeOnly))

			return a
		},
	}

	handler := slog.NewTextHandler(io.MultiWriter(logFileHandler, os.Stdout), opts)

	return slog.New(handler), nil
}

func FlushLog() error {
	if logFileHandler != nil {
		logFileHandler.Sync()
		return logFileHandler.Close()
	}

	return nil
}
