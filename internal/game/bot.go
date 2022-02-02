package game

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger         *zap.Logger
	cfg            config.Config
	dataRepository data.DataRepository
	bm             health.BeltManager
	hr             health.Repository
	tm             town.Manager
	actionChan     chan<- action.Action
}

func NewBot(
	logger *zap.Logger,
	cfg config.Config,
	bm health.BeltManager,
	hr health.Repository,
	tm town.Manager,
	dr data.DataRepository,
	actionChan chan<- action.Action,
) Bot {
	return Bot{
		logger:         logger,
		cfg:            cfg,
		bm:             bm,
		hr:             hr,
		tm:             tm,
		dataRepository: dr,
		actionChan:     actionChan,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	b.prepare()
	//b.tm.Repair(d.Area)
	//helper.NewGame(b.actionChan, b.cfg.Character.Difficulty)
	//// TODO: Check for game creation finished (somehow) instead of waiting for a fixed period of time
	//time.Sleep(time.Second * 10)
	//

	return nil
}

func (b Bot) data() data.Data {
	return b.dataRepository.GameData()
}
