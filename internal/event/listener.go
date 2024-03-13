package event

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/helper"
	"log/slog"
	"os"
	"time"
)

type Listener struct {
	handlers      []Handler
	logger        *slog.Logger
	characterName string
	events        <-chan Event
}

type Handler func(ctx context.Context, e Event) error

func NewListener(logger *slog.Logger, characterName string, channel chan Event) *Listener {
	return &Listener{
		events:        channel,
		characterName: characterName,
		logger:        logger,
	}
}

func (l *Listener) Register(h Handler) {
	l.handlers = append(l.handlers, h)
}

func (l *Listener) Listen(ctx context.Context) error {
	for {
		select {
		case e := <-l.events:
			if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
				err = os.MkdirAll("screenshots", os.ModePerm)
				if err != nil {
					l.logger.Error("error creating screenshots directory", slog.Any("error", err))
				}
			}

			if e.Image() != nil {
				fileName := fmt.Sprintf("screenshots/error-%s-%s.jpeg", l.characterName, time.Now().Format("2006-01-02 15_04_05"))
				err := helper.SaveImageJPEG(e.Image(), fileName)
				if err != nil {
					l.logger.Error("error saving screenshot", slog.Any("error", err))
				}
			}

			for _, h := range l.handlers {
				if err := h(ctx, e); err != nil {
					l.logger.Error("error running event handler", slog.Any("error", err))
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
