package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper/winproc"
	"github.com/hectorgimenez/koolo/internal/run"
	"log/slog"
)

type Supervisor interface {
	Start() error
	Name() string
	Stop()
	TogglePause()
	Stats() Stats
}

type baseSupervisor struct {
	logger       *slog.Logger
	bot          *Bot
	gr           *game.MemoryReader
	gm           *game.Manager
	gi           *game.MemoryInjector
	runFactory   *run.Factory
	name         string
	eventChan    chan<- event.Event
	statsHandler *StatsHandler
	listener     *event.Listener
	cancelFn     context.CancelFunc
}

func newBaseSupervisor(
	logger *slog.Logger,
	bot *Bot,
	gr *game.MemoryReader,
	gm *game.Manager,
	gi *game.MemoryInjector,
	runFactory *run.Factory,
	name string,
	eventChan chan<- event.Event,
	statsHandler *StatsHandler,
	listener *event.Listener,
) (*baseSupervisor, error) {
	return &baseSupervisor{
		logger:       logger,
		bot:          bot,
		gr:           gr,
		gm:           gm,
		gi:           gi,
		runFactory:   runFactory,
		name:         name,
		eventChan:    eventChan,
		statsHandler: statsHandler,
		listener:     listener,
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
		s.gi.RestoreMemory()
	} else {
		s.statsHandler.SetStatus(InGame)
		s.gi.Load()
	}
}

func (s *baseSupervisor) Stop() {
	s.logger.Info("Stopping...", slog.String("configuration", s.name))
	if s.cancelFn != nil {
		s.cancelFn()
	}

	s.gi.Unload()
	s.logger.Info("Finished stopping", slog.String("configuration", s.name))
}

func (s *baseSupervisor) ensureProcessIsRunningAndPrepare(ctx context.Context) error {
	// Prevent screen from turning off
	winproc.SetThreadExecutionState.Call(winproc.EXECUTION_STATE_ES_DISPLAY_REQUIRED | winproc.EXECUTION_STATE_ES_CONTINUOUS)

	// TODO: refactor this
	go s.listener.Listen(ctx)

	return s.gi.Load()
}

func (s *baseSupervisor) logGameStart(runs []run.Run) {
	runNames := ""
	for _, r := range runs {
		runNames += r.Name() + ", "
	}
	s.logger.Info(fmt.Sprintf("Starting Game #%d. Run list: %s", s.statsHandler.Stats().TotalGames(), runNames[:len(runNames)-2]))
}
