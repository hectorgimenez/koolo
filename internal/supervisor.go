package koolo

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/stats"
	"github.com/lxn/win"
	"go.uber.org/zap"
	"os"
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

	firstRun := true
	for {
		if err = helper.NewGame(ctx); err != nil {
			s.logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
			continue
		}

		gameStart := time.Now()
		s.logGameStart(runs)
		err = s.bot.Run(ctx, firstRun, runs)
		if exitErr := helper.ExitGame(ctx); exitErr != nil {
			s.logger.Fatal(fmt.Sprintf("Error exiting game: %s, shutting down...", exitErr))
		}
		firstRun = false

		gameDuration := time.Since(gameStart)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), gameDuration.Seconds()))
		}
		s.updateGameStats()
		helper.Sleep(10000)
	}

}

func (s *Supervisor) updateGameStats() {
	if _, err := os.Stat("stats"); os.IsNotExist(err) {
		err = os.MkdirAll("stats", 0700)
		if err != nil {
			s.logger.Error("Error creating stats directory", zap.Error(err))
			return
		}
	}

	fileName := fmt.Sprintf("stats/stats_%s.txt", stats.Status.ApplicationStartedAt.Format("2006-02-01-15_04_05"))
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Error("Error writing game stats", zap.Error(err))
		return
	}
	w := bufio.NewWriter(f)

	for runName, rs := range stats.Status.RunStats {
		var items = ""
		for _, item := range rs.ItemsFound {
			items += fmt.Sprintf("%s [%s]\n", item.Name, item.Quality)
		}
		avgRunTime := rs.TotalRunsTime.Seconds() / float64(rs.Errors+rs.Kills+rs.Deaths+rs.Chickens+rs.MerChicken)
		statsRun := fmt.Sprintf("Stats for: %s\n"+
			"    Run time: %0.2fs (Total) %0.2fs (Average)\n"+
			"    Kills: %d\n"+
			"    Deaths: %d\n"+
			"    Chickens: %d\n"+
			"    Merc Chickens: %d\n"+
			"    Errors: %d\n"+
			"    Used HP Potions: %d\n"+
			"    Used MP Potions: %d\n"+
			"    Used Rejuv Potions: %d\n"+
			"    Used Merc HP Potions: %d\n"+
			"    Used Merc Rejuv Potions: %d\n"+
			"    Items: \n"+
			"    %s",
			runName,
			rs.TotalRunsTime.Seconds(), avgRunTime,
			rs.Kills,
			rs.Deaths,
			rs.Chickens,
			rs.MerChicken,
			rs.Errors,
			rs.HealingPotionsUsed,
			rs.ManaPotionsUsed,
			rs.RejuvPotionsUsed,
			rs.MercHealingPotionsUsed,
			rs.MercRejuvPotionsUsed,
			items,
		)
		_, err = w.WriteString(statsRun + "\n")
		if err != nil {
			s.logger.Error("Error writing stats file", zap.Error(err))
		}
	}

	w.Flush()
	f.Close()
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

	s.logger.Info(fmt.Sprintf(
		"Diablo II: Resurrected window detected, offsetX: %d offsetY: %d. Game Area Size X: %d Y: %d",
		hid.WindowLeftX,
		hid.WindowTopY,
		hid.GameAreaSizeX,
		hid.GameAreaSizeY,
	))

	stats.Status.ApplicationStartedAt = time.Now()
	return nil
}

func (s *Supervisor) logGameStart(runs []run.Run) {
	runNames := ""
	for _, r := range runs {
		runNames += r.Name() + ", "
	}
	stats.Status.TotalGames++
	s.logger.Info(fmt.Sprintf("Starting Game #%d. Run list: %s", stats.Status.TotalGames, runNames[:len(runNames)-2]))
}
