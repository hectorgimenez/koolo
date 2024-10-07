package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ct "github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/utils/winproc"
	"github.com/lxn/win"
)

type Supervisor interface {
	Start() error
	Name() string
	Stop()
	Stats() Stats
	TogglePause()
	SetWindowPosition(x, y int)
	GetData() *game.Data
	GetContext() *ct.Context
}

type baseSupervisor struct {
	bot          *Bot
	name         string
	statsHandler *StatsHandler
	cancelFn     context.CancelFunc
}

func newBaseSupervisor(
	bot *Bot,
	name string,
	statsHandler *StatsHandler,
) (*baseSupervisor, error) {
	return &baseSupervisor{
		bot:          bot,
		name:         name,
		statsHandler: statsHandler,
	}, nil
}

func (s *baseSupervisor) Name() string {
	return s.name
}

func (s *baseSupervisor) Stats() Stats {
	return s.statsHandler.Stats()
}

func (s *baseSupervisor) TogglePause() {
	if s.bot.ctx.ExecutionPriority == ct.PriorityPause {
		s.bot.ctx.MemoryInjector.Load()
		s.bot.ctx.SwitchPriority(ct.PriorityNormal)
		s.bot.ctx.Logger.Info("Resuming...", slog.String("configuration", s.name))
		event.Send(event.GamePaused(event.Text(s.name, "Game resumed"), false))
		return
	}
	s.bot.ctx.SwitchPriority(ct.PriorityPause)
	s.bot.ctx.MemoryInjector.RestoreMemory()
	s.bot.ctx.Logger.Info("Pausing...", slog.String("configuration", s.name))
	event.Send(event.GamePaused(event.Text(s.name, "Game paused"), true))
}

func (s *baseSupervisor) Stop() {
	s.bot.ctx.Logger.Info("Stopping...", slog.String("configuration", s.name))
	if s.cancelFn != nil {
		s.cancelFn()
	}

	s.bot.ctx.SwitchPriority(ct.PriorityStop)

	s.bot.ctx.MemoryInjector.Unload()
	s.bot.ctx.GameReader.Close()

	if s.bot.ctx.CharacterCfg.KillD2OnStop || s.bot.ctx.CharacterCfg.Scheduler.Enabled {
		s.KillClient()
	}

	s.bot.ctx.Logger.Info("Finished stopping", slog.String("configuration", s.name))
}

func (s *baseSupervisor) KillClient() error {

	process, err := os.FindProcess(int(s.bot.ctx.GameReader.Process.GetPID()))
	if err != nil {
		s.bot.ctx.Logger.Info("Failed to find process", slog.String("configuration", s.name))
		return err
	}
	err = process.Kill()
	if err != nil {
		s.bot.ctx.Logger.Info("Failed to kill process", slog.String("configuration", s.name))
		return err
	}
	return nil
}

func (s *baseSupervisor) ensureProcessIsRunningAndPrepare() error {
	// Prevent screen from turning off
	winproc.SetThreadExecutionState.Call(winproc.EXECUTION_STATE_ES_DISPLAY_REQUIRED | winproc.EXECUTION_STATE_ES_CONTINUOUS)

	return s.bot.ctx.MemoryInjector.Load()
}

func (s *baseSupervisor) logGameStart(runs []run.Run) {
	runNames := ""
	for _, r := range runs {
		runNames += r.Name() + ", "
	}
	s.bot.ctx.Logger.Info(fmt.Sprintf("Starting Game #%d. Run list: %s", s.statsHandler.Stats().TotalGames(), runNames[:len(runNames)-2]))
}

func (s *baseSupervisor) waitUntilCharacterSelectionScreen() error {
	s.bot.ctx.Logger.Info("Waiting for character selection screen...")

	for !s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
		// Spam left click to skip to the char select screen
		s.bot.ctx.HID.Click(game.LeftButton, 100, 100)
		time.Sleep(250 * time.Millisecond)
	}

	s.bot.ctx.Logger.Info("Character selection screen found")

	if s.bot.ctx.CharacterCfg.CharacterName != "" {
		s.bot.ctx.Logger.Info("Selecting character...")
		previousSelection := ""
		for {
			characterName := s.bot.ctx.GameReader.GameReader.GetSelectedCharacterName()
			if strings.EqualFold(previousSelection, characterName) {
				return fmt.Errorf("character %s not found", s.bot.ctx.CharacterCfg.CharacterName)
			}
			if strings.EqualFold(characterName, s.bot.ctx.CharacterCfg.CharacterName) {
				s.bot.ctx.Logger.Info("Character found")
				return nil
			}

			s.bot.ctx.HID.PressKey(win.VK_DOWN)
			time.Sleep(time.Millisecond * 150)
			previousSelection = characterName
		}
	}

	return nil
}

func (s *baseSupervisor) SetWindowPosition(x, y int) {
	uFlags := win.SWP_NOZORDER | win.SWP_NOSIZE | win.SWP_NOACTIVATE
	win.SetWindowPos(s.bot.ctx.GameReader.HWND, 0, int32(x), int32(y), 0, 0, uint32(uFlags))
}
