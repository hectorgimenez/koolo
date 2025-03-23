package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

var logFileHandler *os.File

// FlushLog ensures that all logs are written to file
func FlushLog() error {
	if logFileHandler != nil {
		logFileHandler.Sync()
		return logFileHandler.Close()
	}

	return nil
}

// FlushLogOnly syncs logs to disk without closing the file
func FlushLogOnly() {
	if logFileHandler != nil {
		logFileHandler.Sync()
	}
}

type errorHandler struct {
	slog.Handler
}

// Handle extends the standard handler to flush logs when errors are logged
func (h *errorHandler) Handle(ctx context.Context, record slog.Record) error {
	// Immediately flush logs for error and above to ensure they're written
	// even if the program crashes right after
	if record.Level >= slog.LevelError {
		defer FlushLogOnly()
	}
	return h.Handler.Handle(ctx, record)
}

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

	fileName := "Koolo-log-" + time.Now().Format("2006-01-02-15-04-05") + ".txt"
	if supervisor != "" {
		fileName = fmt.Sprintf("Supervisor-log-%s-%s.txt", supervisor, time.Now().Format("2006-01-02-15-04-05"))
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

	// Create the base handler
	baseHandler := slog.NewTextHandler(io.MultiWriter(logFileHandler, os.Stdout), opts)

	// Wrap it with our error handler that flushes logs on errors
	handler := &errorHandler{Handler: baseHandler}

	return slog.New(handler), nil
}
