package koolo

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/memory"
	"os"
	"syscall"
	"time"

	"github.com/hectorgimenez/koolo/internal/event/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/lxn/win"
	"go.uber.org/zap"
)

type baseSupervisor struct {
	logger *zap.Logger
	bot    *Bot
	gr     *reader.GameReader
	gm     *helper.GameManager
}

func (s *baseSupervisor) TogglePause() {
	s.bot.TogglePause()
}

func (s *baseSupervisor) Stop() {
	s.logger.Info("Shutting down NOW")
	os.Exit(0)
}

func (s *baseSupervisor) updateGameStats() {
	if _, err := os.Stat("stats"); os.IsNotExist(err) {
		err = os.MkdirAll("stats", 0700)
		if err != nil {
			s.logger.Error("Error creating stats directory", zap.Error(err))
			return
		}
	}

	fileName := fmt.Sprintf("stats/stats_%s.txt", stat.Status.ApplicationStartedAt.Format("2006-02-01-15_04_05"))
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Error("Error writing game stats", zap.Error(err))
		return
	}
	w := bufio.NewWriter(f)

	for runName, rs := range stat.Status.RunStats {
		var items = ""
		for _, item := range rs.ItemsFound {
			items += fmt.Sprintf("%s [%d]\n", item.Name, item.Quality)
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

func (s *baseSupervisor) ensureProcessIsRunningAndPrepare() error {
	ptr, err := syscall.UTF16PtrFromString("Diablo II: Resurrected")
	if err != nil {
		return err
	}
	window := win.FindWindow(nil, ptr)
	if window == win.HWND_TOP {
		return errors.New("diablo II: Resurrected window can not be found! Ensure game is open")
	}
	win.SetForegroundWindow(window)
	memory.HWND = window

	pos := win.WINDOWPLACEMENT{}
	point := win.POINT{}
	win.ClientToScreen(window, &point)
	win.GetWindowPlacement(window, &pos)

	hid.WindowLeftX = int(point.X + 1)
	hid.WindowTopY = int(point.Y) + 10
	hid.GameAreaSizeX = int(pos.RcNormalPosition.Right) - hid.WindowLeftX - 9
	hid.GameAreaSizeY = int(pos.RcNormalPosition.Bottom) - hid.WindowTopY - 9
	helper.Sleep(1000)

	s.logger.Info(fmt.Sprintf(
		"Diablo II: Resurrected window detected, offsetX: %d offsetY: %d. Game Area Size X: %d Y: %d",
		hid.WindowLeftX,
		hid.WindowTopY,
		hid.GameAreaSizeX,
		hid.GameAreaSizeY,
	))

	stat.Status.ApplicationStartedAt = time.Now()
	return nil
}

func (s *baseSupervisor) logGameStart(runs []run.Run) {
	runNames := ""
	for _, r := range runs {
		runNames += r.Name() + ", "
	}
	stat.Status.TotalGames++
	s.logger.Info(fmt.Sprintf("Starting Game #%d. Run list: %s", stat.Status.TotalGames, runNames[:len(runNames)-2]))
}
