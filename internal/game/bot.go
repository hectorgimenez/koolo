package game

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/item"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger         *zap.Logger
	cfg            config.Config
	dataRepository data.DataRepository
	bm             health.BeltManager
	tm             town.Manager
	char           character.Character
	runs           []run.Run
	pickup         item.Pickup
}

func NewBot(
	logger *zap.Logger,
	cfg config.Config,
	bm health.BeltManager,
	tm town.Manager,
	dr data.DataRepository,
	char character.Character,
	runs []run.Run,
	pickup item.Pickup,
) Bot {
	return Bot{
		logger:         logger,
		cfg:            cfg,
		bm:             bm,
		tm:             tm,
		dataRepository: dr,
		char:           char,
		runs:           runs,
		pickup:         pickup,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	b.prepare()

	for _, r := range b.runs {
		err := r.MoveToStartingPoint()
		if err != nil {
			// TODO: Handle error
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

	//helper.NewGame(b.actionChan, b.cfg.Character.Difficulty)
	//// TODO: Check for game creation finished (somehow) instead of waiting for a fixed period of time

	return nil
}

func (b Bot) data() data.Data {
	return b.dataRepository.GameData()
}
