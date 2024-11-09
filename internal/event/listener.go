package event

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/utils"
)

var events = make(chan Event)

type Listener struct {
	handlers         []Handler
	deliveryHandlers map[int]Handler
	logger           *slog.Logger
}

type Handler func(ctx context.Context, e Event) error

func NewListener(logger *slog.Logger) *Listener {
	return &Listener{
		logger:           logger,
		deliveryHandlers: make(map[int]Handler),
	}
}

func (l *Listener) Register(h Handler) {
	l.handlers = append(l.handlers, h)
}

func (l *Listener) Listen(ctx context.Context) error {
	for {
		select {
		case e := <-events:
			if _, err := os.Stat("screenshots"); os.IsNotExist(err) {
				err = os.MkdirAll("screenshots", os.ModePerm)
				if err != nil {
					l.logger.Error("error creating screenshots directory", slog.Any("error", err))
				}
			}

			if e.Image() != nil && config.Koolo.Debug.Screenshots {
				fileName := fmt.Sprintf("screenshots/error-%s.jpeg", time.Now().Format("2006-01-02 15_04_05"))
				err := utils.SaveImageJPEG(e.Image(), fileName)
				if err != nil {
					l.logger.Error("error saving screenshot", slog.Any("error", err))
				}
			}

			for _, h := range l.handlers {
				if err := h(ctx, e); err != nil && e.Message() != "" {
					l.logger.Error("error running event handler", slog.Any("error", err))
				}
			}
			for _, h := range l.deliveryHandlers {
				if err := h(ctx, e); err != nil {
					l.logger.Error("error running event delivery handler", slog.Any("error", err))
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (l *Listener) WaitForEvent(ctx context.Context) Event {
	evtChan := make(chan Event)
	idx := rand.Intn(math.MaxInt64)
	l.deliveryHandlers[idx] = func(ctx context.Context, e Event) error {
		evtChan <- e
		return nil
	}
	// Clean up the handler when we're done
	defer func() {
		delete(l.deliveryHandlers, idx)
	}()

	for {
		select {
		case e := <-evtChan:
			return e
		case <-ctx.Done():
			return nil
		}
	}
}

func Send(e Event) {
	events <- e
}
