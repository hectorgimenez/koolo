package koolo

import "context"

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
}

func NewBot() Bot {
	return Bot{}
}

func (b Bot) Start(ctx context.Context) error {
	return nil
}
