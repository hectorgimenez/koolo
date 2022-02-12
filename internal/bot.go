package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/item"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger *zap.Logger
	cfg    config.Config
	bm     health.BeltManager
	tm     town.Manager
	char   character.Character
	runs   []run.Run
	pickup item.Pickup
}

func NewBot(
	logger *zap.Logger,
	cfg config.Config,
	bm health.BeltManager,
	tm town.Manager,
	char character.Character,
	runs []run.Run,
	pickup item.Pickup,
) Bot {
	return Bot{
		logger: logger,
		cfg:    cfg,
		bm:     bm,
		tm:     tm,
		char:   char,
		runs:   runs,
		pickup: pickup,
	}
}

func (b *Bot) Start(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r == game.NotInGameErr {
			err = game.NotInGameErr
		}
	}()

	start := time.Now()

	for i, r := range b.runs {
		if i == 0 {
			b.prepare(true)
		} else {
			b.prepare(false)
		}

		err = r.MoveToStartingPoint()
		if err != nil {
			b.logger.Error("Error moving to start point for current run, let's skip it")
			continue
		}

		err = r.TravelToDestination()
		if err != nil {
			b.quitIfErr(start, b.char.ReturnToTown())
			continue
		}

		err = r.Kill()
		if err != nil {
			b.quitIfErr(start, b.char.ReturnToTown())
			continue
		}
		b.logger.Debug("Run cleared, picking up items...")
		b.pickup.Pickup()

		// Don't return to town on last run, just quit
		if len(b.runs)-1 != i {
			b.logger.Debug("Item pickup completed, returning to town...")
			b.quitIfErr(start, b.char.ReturnToTown())
		}
	}

	b.logger.Info(fmt.Sprintf("Game finished successfully. Run time: %0.2f seconds", time.Since(start).Seconds()))
	helper.ExitGame()

	return nil
}

func (b *Bot) quitIfErr(startTime time.Time, err error) {
	if err != nil {
		helper.ExitGame()
		b.logger.Info(fmt.Sprintf("Game finished with errors. Run time: %0.2f seconds, error: %s", time.Since(startTime).Seconds(), err.Error()))
	}
}
