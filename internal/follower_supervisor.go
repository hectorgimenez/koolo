package koolo

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/run"
	"log/slog"
)

type FollowerSupervisor interface {
	Start(ctx context.Context, factory *run.Factory) error
	Stop()
	TogglePause()
}

type FollowerPlayerSupervisor struct {
	baseSupervisor
}

func NewFollowerPlayerSupervisor(logger *slog.Logger, bot *Bot, gr *reader.GameReader, gm *helper.GameManager) *FollowerPlayerSupervisor {
	return &FollowerPlayerSupervisor{
		baseSupervisor: baseSupervisor{
			logger: logger,
			bot:    bot,
			gr:     gr,
			gm:     gm,
		},
	}
}

func (f *FollowerPlayerSupervisor) Start(ctx context.Context, factory *run.Factory) error {
	err := f.ensureProcessIsRunningAndPrepare()
	if err != nil {
		return fmt.Errorf("error preparing game: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			for {
				if !f.gm.InGame() {
					return fmt.Errorf("must be in game with the leader")
				}

				runs := factory.BuildRuns()

				err := f.bot.Run(ctx, false, runs)
				if err != nil {
					if errors.Is(context.Canceled, ctx.Err()) {
						f.logger.Error("Context canceled")
						return nil
					}
					errorMsg := fmt.Sprintf("Game finished with errors, reason: %s", err.Error())
					event.Events <- event.WithScreenshot(errorMsg)
					f.logger.Warn(errorMsg)
				}

				helper.Sleep(100)
			}
		}
	}
}

func (f *FollowerPlayerSupervisor) JoinGame(gameName, password string) {
	// No need to join game for this proof of concept implementation
}
