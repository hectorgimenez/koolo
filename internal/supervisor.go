package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/container"
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
	cancelFn     context.CancelFunc
	c            container.Container
}

func newBaseSupervisor(
	bot *Bot,
	runFactory *run.Factory,
	name string,
	statsHandler *StatsHandler,
	c container.Container,
) (*baseSupervisor, error) {
	return &baseSupervisor{
		bot:          bot,
		runFactory:   runFactory,
		name:         name,
		statsHandler: statsHandler,
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

	return s.c.Injector.Load()
}

func (s *baseSupervisor) logGameStart(runs []run.Run) {
	runNames := ""
	for _, r := range runs {
		runNames += r.Name() + ", "
	}
	s.c.Logger.Info(fmt.Sprintf("Starting Game #%d. Run list: %s", s.statsHandler.Stats().TotalGames(), runNames[:len(runNames)-2]))
}

func (s *baseSupervisor) waitUntilCharacterSelectionScreen() error {
	s.c.Logger.Info("Waiting for character selection screen...")

	for s.c.Reader.GameReader.GetSelectedCharacterName() == "" {
		s.c.HID.Click(game.LeftButton, 100, 100)
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second) // Add an extra second to allow UI to properly render on slow computers

	s.c.Logger.Info("Character selection screen found")

	if s.c.CharacterCfg.CharacterName != "" {
		s.c.Logger.Info("Selecting character...")
		previousSelection := ""
		for {
			characterName := s.c.Reader.GameReader.GetSelectedCharacterName()
			if previousSelection == characterName {
				return fmt.Errorf("character %s not found", s.c.CharacterCfg.CharacterName)
			}
			if characterName == s.c.CharacterCfg.CharacterName {
				s.c.Logger.Info("Character found")
				return nil
			}

			s.c.HID.PressKey("down")
			time.Sleep(time.Millisecond * 500)
			previousSelection = characterName
		}
	}

	return nil
}
