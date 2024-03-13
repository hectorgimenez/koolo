package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper/winproc"
	"github.com/hectorgimenez/koolo/internal/run"
	"log/slog"
	"time"
)

type Supervisor interface {
	Start() error
	Name() string
	Stop()
	TogglePause()
	Stats() Stats
}

type baseSupervisor struct {
	bot          *Bot
	runFactory   *run.Factory
	name         string
	statsHandler *StatsHandler
	listener     *event.Listener
	cancelFn     context.CancelFunc
	c            container.Container
}

func newBaseSupervisor(
	bot *Bot,
	runFactory *run.Factory,
	name string,
	statsHandler *StatsHandler,
	listener *event.Listener,
	c container.Container,
) (*baseSupervisor, error) {
	return &baseSupervisor{
		bot:          bot,
		runFactory:   runFactory,
		name:         name,
		statsHandler: statsHandler,
		listener:     listener,
		c:            c,
	}, nil
}

func (s *baseSupervisor) Name() string {
	return s.name
}

func (s *baseSupervisor) Stats() Stats {
	return s.statsHandler.Stats()
}

func (s *baseSupervisor) TogglePause() {
	s.bot.TogglePause()
	if s.bot.paused {
		s.statsHandler.SetStatus(Paused)
		s.c.Injector.RestoreMemory()
	} else {
		s.statsHandler.SetStatus(InGame)
		s.c.Injector.Load()
	}
}

func (s *baseSupervisor) Stop() {
	s.c.Logger.Info("Stopping...", slog.String("configuration", s.name))
	if s.cancelFn != nil {
		s.cancelFn()
	}

	s.c.Injector.Unload()
	s.c.Logger.Info("Finished stopping", slog.String("configuration", s.name))
}

func (s *baseSupervisor) ensureProcessIsRunningAndPrepare(ctx context.Context) error {
	// Prevent screen from turning off
	winproc.SetThreadExecutionState.Call(winproc.EXECUTION_STATE_ES_DISPLAY_REQUIRED | winproc.EXECUTION_STATE_ES_CONTINUOUS)

	// TODO: refactor this
	go s.listener.Listen(ctx)

	err := s.c.Injector.Load()
	if err != nil {
		return err
	}

	s.c.Logger.Info("Waiting for character selection screen...")
	for range 25 {
		s.c.HID.Click(game.LeftButton, 100, 100)
		time.Sleep(time.Second)
	}

	s.c.Logger.Info("Trying to start game...")

	return nil
}

func (s *baseSupervisor) logGameStart(runs []run.Run) {
	runNames := ""
	for _, r := range runs {
		runNames += r.Name() + ", "
	}
	s.c.Logger.Info(fmt.Sprintf("Starting Game #%d. Run list: %s", s.statsHandler.Stats().TotalGames(), runNames[:len(runNames)-2]))
}
