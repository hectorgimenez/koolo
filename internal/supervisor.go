package koolo

import (
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/lxn/win"
	"go.uber.org/zap"
	"time"
)

// Supervisor is the main bot entrypoint, it will handle all the parallel processes and ensure everything is up and running
type Supervisor struct {
	logger         *zap.Logger
	cfg            config.Config
	eventsChannel  <-chan event.Event
	healthManager  health.Manager
	bot            Bot
	executionStats ExecutionStats
}

type ExecutionStats struct {
	Runs        map[string]*RunStats
	RunningTime time.Duration
}

type RunStats struct {
	ItemCounter int
	Kills       int
	Deaths      int
	Chickens    int
	Errors      int
	Time        time.Duration
}

func NewSupervisor(logger *zap.Logger, cfg config.Config, hm health.Manager, bot Bot) Supervisor {
	return Supervisor{
		logger:        logger,
		cfg:           cfg,
		healthManager: hm,
		bot:           bot,
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s *Supervisor) Start(ctx context.Context, runs []run.Run) error {
	s.initializeStats(runs)
	err := s.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	for {
		s.logger.Info("Creating new game...")
		err = helper.NewGame(s.cfg.Character.Difficulty)
		if err != nil {
			s.logger.Fatal("Error creating new game")
		}

		gameStats := s.bot.RunGame(ctx, runs)
		for k, stats := range gameStats {
			s.executionStats.Runs[k].ItemCounter = s.executionStats.Runs[k].ItemCounter + stats.ItemCounter
			s.executionStats.Runs[k].Kills = s.executionStats.Runs[k].Kills + stats.Kills
			s.executionStats.Runs[k].Deaths = s.executionStats.Runs[k].Deaths + stats.Deaths
			s.executionStats.Runs[k].Chickens = s.executionStats.Runs[k].Chickens + stats.Chickens
			s.executionStats.Runs[k].Errors = s.executionStats.Runs[k].Errors + stats.Errors
			s.executionStats.Runs[k].Time = s.executionStats.Runs[k].Time + stats.Time
		}

		time.Sleep(time.Second * 5)
	}
}

func (s *Supervisor) ensureProcessIsRunningAndPrepare() error {
	window := robotgo.FindWindow("Diablo II: Resurrected")
	if window == win.HWND_TOP {
		s.logger.Fatal("Diablo II: Resurrected window can not be found! Are you sure game is open?")
	}
	win.SetForegroundWindow(window)

	// Exclude border offsets
	// TODO: Improve this, maybe getting window content coordinates?
	pos := win.WINDOWPLACEMENT{}
	win.GetWindowPlacement(window, &pos)
	hid.WindowLeftX = int(pos.RcNormalPosition.Left) + 8
	hid.WindowTopY = int(pos.RcNormalPosition.Top) + 31
	hid.GameAreaSizeX = int(pos.RcNormalPosition.Right) - hid.WindowLeftX - 10
	hid.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - hid.WindowTopY - 10
	time.Sleep(time.Second * 1)

	return nil
}

func (s *Supervisor) initializeStats(runs []run.Run) {
	s.executionStats.Runs = map[string]*RunStats{}
	for _, r := range runs {
		s.executionStats.Runs[r.Name()] = &RunStats{}
	}
}
