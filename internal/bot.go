package koolo

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/item"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
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

func (b *Bot) Start(ctx context.Context) error {
	b.prepare()

	for _, r := range b.runs {
		err := r.MoveToStartingPoint()
		if err != nil {
			b.logger.Error("Error moving to start point for current run, let's skip it")
			continue
		}

		err = r.TravelToDestination()
		if err != nil {
			r.ReturnToTown()
			continue
		}

		err = r.Kill()
		if err != nil {
			r.ReturnToTown()
			continue
		}
		b.logger.Debug("Run cleared, picking up items...")
		b.pickup.Pickup()

		b.logger.Debug("Item pickup completed, returning to town...")
		r.ReturnToTown()
	}

	return nil
}
