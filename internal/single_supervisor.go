package koolo

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/run"
	"log/slog"
	"math/rand"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

type SinglePlayerSupervisor struct {
	*baseSupervisor
}

func NewSinglePlayerSupervisor(name string, logger *slog.Logger, bot *Bot, gr *game.MemoryReader, gm *game.Manager, gi *game.MemoryInjector, runFactory *run.Factory, eventChan chan<- event.Event, statsHandler *StatsHandler, listener *event.Listener) (*SinglePlayerSupervisor, error) {
	bs, err := newBaseSupervisor(logger, bot, gr, gm, gi, runFactory, name, eventChan, statsHandler, listener)
	if err != nil {
		return nil, err
	}

	return &SinglePlayerSupervisor{
		baseSupervisor: bs,
	}, nil
}

// Start will return error if it can not be started, otherwise will always return nil
func (s *SinglePlayerSupervisor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	err := s.ensureProcessIsRunningAndPrepare(ctx)
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	firstRun := true
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if !s.gm.InGame() {
					if err = s.gm.NewGame(); err != nil {
						s.logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
						continue
					}
				}

				runs := s.runFactory.BuildRuns()
				gameStart := time.Now()
				if config.Config.Game.RandomizeRuns {
					rand.Shuffle(len(runs), func(i, j int) { runs[i], runs[j] = runs[j], runs[i] })
				}
				s.logGameStart(runs)
				err = s.bot.Run(ctx, firstRun, runs)
				if err != nil {
					if errors.Is(context.Canceled, ctx.Err()) {
						continue
					}

					switch {
					case errors.Is(err, health.ErrChicken):
						s.eventChan <- event.GameFinished(event.WithScreenshot(err.Error(), s.gr.Screenshot()), event.FinishedChicken)
						s.logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
					case errors.Is(err, health.ErrMercChicken):
						s.eventChan <- event.GameFinished(event.WithScreenshot(err.Error(), s.gr.Screenshot()), event.FinishedMercChicken)
						s.logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
					case errors.Is(err, health.ErrDied):
						s.eventChan <- event.GameFinished(event.WithScreenshot(err.Error(), s.gr.Screenshot()), event.FinishedDied)
						s.logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
					default:
						s.eventChan <- event.GameFinished(event.WithScreenshot(err.Error(), s.gr.Screenshot()), event.FinishedError)
						s.logger.Warn(fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds()))
					}
				}
				if exitErr := s.gm.ExitGame(); exitErr != nil {
					errMsg := fmt.Sprintf("Error exiting game %s", err.Error())
					s.eventChan <- event.GameFinished(event.WithScreenshot(errMsg, s.gr.Screenshot()), event.FinishedError)
					s.logger.Warn(errMsg)
					return
				}
				firstRun = false
			}
		}
	}()

	return nil
}
