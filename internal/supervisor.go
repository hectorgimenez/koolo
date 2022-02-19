package koolo

import (
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/lxn/win"
	"go.uber.org/zap"
	"time"
)

// Supervisor is the main bot entrypoint, it will handle all the parallel processes and ensure everything is up and running
type Supervisor struct {
	logger *zap.Logger
	bot    Bot
}

type RunStats struct {
	ItemCounter int
	Kills       int
	Deaths      int
	Chickens    int
	Errors      int
	Time        time.Duration
}

func NewSupervisor(logger *zap.Logger, bot Bot) Supervisor {
	return Supervisor{
		logger: logger,
		bot:    bot,
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s *Supervisor) Start(ctx context.Context, runs []run.Run) error {
	err := s.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	return s.bot.Run(ctx, runs)
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
