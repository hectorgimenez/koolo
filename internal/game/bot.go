package game

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/action"
	"go.uber.org/zap"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger         *zap.Logger
	dataRepository DataRepository
	gm             GameManager
	actionChan     chan<- action.Action
}

func NewBot(logger *zap.Logger, gm GameManager, dr DataRepository, actionChan chan<- action.Action) Bot {
	return Bot{
		logger:         logger,
		gm:             gm,
		dataRepository: dr,
		actionChan:     actionChan,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	b.gm.NewGame()
	time.Sleep(time.Second * 10)
	b.prepare()
	return nil
}
