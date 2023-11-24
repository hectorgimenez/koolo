package koolo

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/run"
	"go.uber.org/zap"
)

type Supervisor interface {
	Start(ctx context.Context, factory *run.Factory) error
	Stop()
	TogglePause()
}

type SinglePlayerSupervisor struct {
	baseSupervisor
}

func NewSinglePlayerSupervisor(logger *zap.Logger, bot *Bot, gr *reader.GameReader, gm *helper.GameManager) *SinglePlayerSupervisor {
	return &SinglePlayerSupervisor{
		baseSupervisor: baseSupervisor{
			logger: logger,
			bot:    bot,
			gr:     gr,
			gm:     gm,
		},
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s *SinglePlayerSupervisor) Start(ctx context.Context, factory *run.Factory) error {
	err := s.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	firstRun := true
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if !s.gm.InGame() {
				if err = s.gm.NewGame(); err != nil {
					s.logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
					continue
				}
			}

			runs := factory.BuildRuns()
			gameStart := time.Now()
			if config.Config.Game.RandomizeRuns {
				rand.Shuffle(len(runs), func(i, j int) { runs[i], runs[j] = runs[j], runs[i] })
			}
			s.logGameStart(runs)
			err = s.bot.Run(ctx, firstRun, runs)
			if err != nil {
				if errors.Is(context.Canceled, ctx.Err()) {
					return nil
				}
				errorMsg := fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds())
				event.Events <- event.WithScreenshot(errorMsg)
				s.logger.Warn(errorMsg)
			}
			if exitErr := s.gm.ExitGame(); exitErr != nil {
				return fmt.Errorf("error exiting game: %s", exitErr)
			}
			firstRun = false

			s.updateGameStats()
		}
	}
}
