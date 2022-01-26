package koolo

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	gm game.GameManager
}

func NewBot(gm game.GameManager) Bot {
	return Bot{
		gm: gm,
	}
}

func (b Bot) Start(ctx context.Context) error {
	b.gm.NewGame()
	time.Sleep(time.Second * 8)
	b.gm.ExitGame()
	return nil
}
