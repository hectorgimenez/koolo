package game

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"go.uber.org/zap"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger         *zap.Logger
	cfg            config.Config
	dataRepository DataRepository
	bm             health.BeltManager
	actionChan     chan<- action.Action
}

func NewBot(logger *zap.Logger, cfg config.Config, bm health.BeltManager, dr DataRepository, actionChan chan<- action.Action) Bot {
	return Bot{
		logger:         logger,
		cfg:            cfg,
		bm:             bm,
		dataRepository: dr,
		actionChan:     actionChan,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	helper.NewGame(b.actionChan, b.cfg.Character.Difficulty)
	// TODO: Check for game creation finished (somehow) instead of waiting for a fixed period of time
	time.Sleep(time.Second * 10)

	b.prepare()
	return nil
}
