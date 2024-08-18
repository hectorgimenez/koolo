package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/run"
	"github.com/hectorgimenez/koolo/internal/v2/utils/winproc"
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
	//s.bot.TogglePause()
}

func (s *baseSupervisor) Stop() {
	s.bot.ctx.Logger.Info("Stopping...", slog.String("configuration", s.name))
	if s.cancelFn != nil {
		s.cancelFn()
	}

	s.bot.ctx.MemoryInjector.Unload()
	s.bot.ctx.GameReader.Close()

	if s.bot.ctx.CharacterCfg.KillD2OnStop {
		process, err := os.FindProcess(int(s.bot.ctx.GameReader.Process.GetPID()))
		if err != nil {
			s.bot.ctx.Logger.Info("Failed to find process", slog.String("configuration", s.name))
		}
		err = process.Kill()
		if err != nil {
			s.bot.ctx.Logger.Info("Failed to kill process", slog.String("configuration", s.name))
		}
	}
	s.bot.ctx.Logger.Info("Finished stopping", slog.String("configuration", s.name))
}

func (s *baseSupervisor) ensureProcessIsRunningAndPrepare(ctx context.Context) error {
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

	for s.bot.ctx.GameReader.GameReader.GetSelectedCharacterName() == "" {
		s.bot.ctx.HID.Click(game.LeftButton, 100, 100)
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second) // Add an extra second to allow UI to properly render on slow computers

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
			time.Sleep(time.Millisecond * 500)
			previousSelection = characterName
		}
	}

	return nil
}

func (s *baseSupervisor) SetWindowPosition(x, y int) {
	uFlags := win.SWP_NOZORDER | win.SWP_NOSIZE | win.SWP_NOACTIVATE
	win.SetWindowPos(s.bot.ctx.GameReader.HWND, 0, int32(x), int32(y), 0, 0, uint32(uFlags))
}
