package bot

import (
	"context"
	"errors"
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
	} else {
		s.bot.ctx.SwitchPriority(ct.PriorityPause)
		s.bot.ctx.MemoryInjector.RestoreMemory()
		s.bot.ctx.Logger.Info("Pausing...", slog.String("configuration", s.name))
		event.Send(event.GamePaused(event.Text(s.name, "Game paused"), true))
	}
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
	disconnected := false

	if err := s.ensureOnline(); err != nil {
		s.bot.ctx.Logger.Error("[Ensure Online]: Failed to prepare for character selection, will kill client ...")
		if err := s.KillClient(); err != nil {
			s.bot.ctx.Logger.Error("[Ensure Online]: Failed to kill client", slog.String("error", err.Error()))
			return err
		}
		return err
	}

	if s.bot.ctx.CharacterCfg.CharacterName != "" {

		s.bot.ctx.Logger.Info("Selecting character...")

		// If we've lost connection it bugs out and we need to select another character and the first one again.
		if disconnected {
			s.bot.ctx.HID.PressKey(win.VK_DOWN)
			time.Sleep(250 * time.Millisecond)
			s.bot.ctx.HID.PressKey(win.VK_UP)
		}

		// Try to select a character up to 25 times then give up and kill the client
		for i := 0; i < 25; i++ {
			characterName := s.bot.ctx.GameReader.GameReader.GetSelectedCharacterName()

			s.bot.ctx.Logger.Debug(fmt.Sprintf("Checking character: %s", characterName))

			if strings.EqualFold(characterName, s.bot.ctx.CharacterCfg.CharacterName) {
				s.bot.ctx.Logger.Info(fmt.Sprintf("Character %s found and selected.", s.bot.ctx.CharacterCfg.CharacterName))
				return nil
			}

			s.bot.ctx.HID.PressKey(win.VK_DOWN)
			time.Sleep(250 * time.Millisecond)
		}

		s.bot.ctx.Logger.Info(fmt.Sprintf("Character %s not found after 25 attempts, terminating client ...", s.bot.ctx.CharacterCfg.CharacterName))

		if err := s.KillClient(); err != nil {
			return err
		}
	}

	return nil
}

func (s *baseSupervisor) SetWindowPosition(x, y int) {
	uFlags := win.SWP_NOZORDER | win.SWP_NOSIZE | win.SWP_NOACTIVATE
	win.SetWindowPos(s.bot.ctx.GameReader.HWND, 0, int32(x), int32(y), 0, 0, uint32(uFlags))
}

func (s *baseSupervisor) ensureOnline() error {
	if !s.bot.ctx.GameReader.IsInCharacterSelectionScreen() {
		return fmt.Errorf("[Ensure Online]: We're not in the character selection screen")
	}

	if !s.bot.ctx.GameReader.IsOnline() && s.bot.ctx.CharacterCfg.AuthMethod != "None" {
		s.bot.ctx.HID.Click(game.LeftButton, 1090, 32)
		s.bot.ctx.Logger.Debug("[Ensure Online]: We're at the character selection screen but not online")

		time.Sleep(2000)

		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			s.bot.ctx.Logger.Debug(fmt.Sprintf("[Ensure Online]: Trying to connect to bnet attempt %d of %d", i+1, maxRetries))
			s.bot.ctx.HID.Click(game.LeftButton, 1090, 32)
			time.Sleep(2000)

			for {
				blockingPanel := s.bot.ctx.GameReader.GetPanel("BlockingPanel")
				popuPanel := s.bot.ctx.GameReader.GetPanel("DismissableModal")

				if blockingPanel.PanelName != "" && blockingPanel.PanelEnabled && blockingPanel.PanelVisible {
					s.bot.ctx.Logger.Debug("[Ensure Online]: Loading panel detected, waiting for it to disappear")
					time.Sleep(2000)
					continue
				}

				if popuPanel.PanelName != "" && popuPanel.PanelEnabled && popuPanel.PanelVisible {
					s.bot.ctx.Logger.Debug("[Ensure Online]: Dismissable modal detected, dismissing it and trying to connect again ...")
					s.bot.ctx.HID.PressKey(0x1B)
					time.Sleep(1000)
					break
				}

				s.bot.ctx.Logger.Debug("[Ensure Online]: No loading panel or errors detected, checking if we're online ...")
				break
			}

			if s.bot.ctx.GameReader.IsOnline() {
				s.bot.ctx.Logger.Debug("[Ensure Online]: We're online!")
				return nil
			}
		}
	} else {
		s.bot.ctx.Logger.Debug("[Ensure Online]: We're already online or don't have to be, skipping ...")
		return nil
	}

	return errors.New("[Ensure Online]: Failed to connect to bnet")
}
