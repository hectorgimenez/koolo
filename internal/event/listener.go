package event

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/helper"
	"os"
	"time"

	"go.uber.org/zap"
)

type Listener struct {
	handlers []Handler
	logger   *zap.Logger
}

type Handler func(ctx context.Context, m Message) error

func NewListener(logger *zap.Logger) Listener {
	return Listener{
		logger: logger,
	}
}

func (l *Listener) Register(h Handler) {
	l.handlers = append(l.handlers, h)
}

func (l *Listener) Listen(ctx context.Context) error {
	for {
		select {
		case e := <-Events:
			if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
				err = os.MkdirAll("screenshots", 0700)
				if err != nil {
					l.logger.Error("error creating screenshots directory", zap.Error(err))
				}
			}

			if e.Image != nil {
				fileName := fmt.Sprintf("screenshots/error-%s.png", time.Now().Format("2006-01-02 15_04_05"))
				err := helper.SavePNG(e.Image, fileName)
				if err != nil {
					l.logger.Error("error saving screenshot", zap.Error(err))
				}
			}

			for _, h := range l.handlers {
				if err := h(ctx, e); err != nil {
					l.logger.Error("error running event handler", zap.Error(err))
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
