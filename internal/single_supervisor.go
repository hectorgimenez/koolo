package koolo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/hectorgimenez/koolo/internal/action"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/run"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
)

type SinglePlayerSupervisor struct {
	*baseSupervisor
}

func NewSinglePlayerSupervisor(name string, bot *Bot, runFactory *run.Factory, statsHandler *StatsHandler, c container.Container, pid uint32, hwnd uintptr) (*SinglePlayerSupervisor, error) {
	bs, err := newBaseSupervisor(bot, runFactory, name, statsHandler, c)
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
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if firstRun {
				err = s.waitUntilCharacterSelectionScreen()
				if err != nil {
					return fmt.Errorf("error waiting for character selection screen: %w", err)
				}
			}
			if !s.c.Manager.InGame() {
				if err = s.c.Manager.NewGame(); err != nil {
					s.c.Logger.Error(fmt.Sprintf("Error creating new game: %s", err.Error()))
					continue
				}
			}

			runs := s.runFactory.BuildRuns()
			gameStart := time.Now()
			if config.Characters[s.name].Game.RandomizeRuns {
				rand.Shuffle(len(runs), func(i, j int) { runs[i], runs[j] = runs[j], runs[i] })
			}
			if config.Koolo.Discord.EnableGameCreatedMessages {
				event.Send(event.GameCreated(event.Text(s.name, "New game created"), "", ""))
			} else {
				event.Send(event.GameCreated(event.Text(s.name, ""), "", ""))
			}
			action.ResetBuffTime(s.Name())
			s.logGameStart(runs)
			err = s.bot.Run(ctx, firstRun, runs)
			if err != nil {
				if errors.Is(context.Canceled, ctx.Err()) {
					continue
				}

				switch {
				case errors.Is(err, health.ErrChicken):
					if config.Koolo.Discord.EnableDiscordChickenMessages {
						event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedChicken))
					} else {
						event.Send(event.GameFinished(event.Text(s.name, ""), event.FinishedChicken))
					}
					s.c.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				case errors.Is(err, health.ErrMercChicken):
					if config.Koolo.Discord.EnableDiscordChickenMessages {
						event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedMercChicken))
					} else {
						event.Send(event.GameFinished(event.Text(s.name, ""), event.FinishedMercChicken))
					}
					s.c.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				case errors.Is(err, health.ErrDied):
					if config.Koolo.Discord.EnableDiscordChickenMessages {
						event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedDied))
					} else {
						event.Send(event.GameFinished(event.Text(s.name, ""), event.FinishedDied))
					}
					s.c.Logger.Warn(err.Error(), slog.Float64("gameLength", time.Since(gameStart).Seconds()))
				default:
					event.Send(event.GameFinished(event.WithScreenshot(s.name, err.Error(), s.c.Reader.Screenshot()), event.FinishedError))
					s.c.Logger.Warn(
						fmt.Sprintf("Game finished with errors, reason: %s. Game total time: %0.2fs", err.Error(), time.Since(gameStart).Seconds()),
						slog.String("supervisor", s.name),
						slog.Uint64("mapSeed", uint64(s.c.Reader.CachedMapSeed)),
					)
				}
			}
			if exitErr := s.c.Manager.ExitGame(); exitErr != nil {
				errMsg := fmt.Sprintf("Error exiting game %s", err.Error())
				event.Send(event.GameFinished(event.WithScreenshot(s.name, errMsg, s.c.Reader.Screenshot()), event.FinishedError))

				return errors.New(errMsg)
			}
			firstRun = false
		}
	}
}
