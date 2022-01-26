package koolo

import (
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/lxn/win"
	"golang.org/x/sync/errgroup"
)

// Supervisor is the main bot entrypoint, it will handle all the parallel processes and ensure everything is up and running
type Supervisor struct {
	cfg           config.Config
	eventsChannel <-chan event.Event
	ah            action.Handler
	healthManager health.Manager
	bot           Bot
}

func NewSupervisor(cfg config.Config, ah action.Handler, hm health.Manager, bot Bot) Supervisor {
	return Supervisor{
		cfg:           cfg,
		ah:            ah,
		healthManager: hm,
		bot:           bot,
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s Supervisor) Start(ctx context.Context) error {
	err := s.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	// Listen to actions triggered from elsewhere
	g.Go(func() error {
		return s.ah.Listen(ctx)
	})

	// Listen to events and attaching/detaching bot operations
	g.Go(func() error {
		s.listenEvents(ctx)
		return nil
	})

	// Main loop will be inside this, will handle bosses and path traveling
	g.Go(func() error {
		return s.bot.Start(ctx)
	})

	// Will keep our character and mercenary alive, monitoring life and mana
	g.Go(func() error {
		return s.healthManager.Start(ctx)
	})

	return g.Wait()
}

func (s Supervisor) listenEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-s.eventsChannel:
			switch evt {
			case event.ExitedGame:
				s.healthManager.Pause()
			case event.SafeAreaAbandoned:
				s.healthManager.Resume()
			case event.SafeAreaEntered:
				s.healthManager.Pause()
			}
		}
	}
}

func (s Supervisor) ensureProcessIsRunningAndPrepare() error {
	window := robotgo.FindWindow("Diablo II: Resurrected")
	win.SetForegroundWindow(window)

	pos := win.WINDOWPLACEMENT{}
	win.GetWindowPlacement(window, &pos)

	// Assuming game resolution 1280x720
	offsetX := ((pos.RcNormalPosition.Right - pos.RcNormalPosition.Left - 1280) / 2) + pos.RcNormalPosition.Left
	offsetY := ((pos.RcNormalPosition.Bottom - pos.RcNormalPosition.Top - 720) / 2) + pos.RcNormalPosition.Top
	hid.WindowPositionX = int(offsetX)
	hid.WindowPositionY = int(offsetY)

	return nil
}
