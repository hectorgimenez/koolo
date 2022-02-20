package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/run"
	"go.uber.org/zap"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger *zap.Logger
	hm     health.Manager
	ab     action.Builder
}

func NewBot(
	logger *zap.Logger,
	hm health.Manager,
	ab action.Builder,
) Bot {
	return Bot{
		logger: logger,
		hm:     hm,
		ab:     ab,
	}
}

func (b *Bot) Run(ctx context.Context, runs []run.Run) error {
	for _, r := range runs {
		runStart := time.Now()
		b.logger.Debug(fmt.Sprintf("Running: %s", r.Name()))

		actions := []action.Action{
			b.ab.RecoverCorpse(),
			b.ab.VendorRefill(),
			b.ab.ReviveMerc(),
			b.ab.Repair(),
			b.ab.Heal(),
			b.ab.Stash(),
		}

		actions = append(actions, r.BuildActions()...)
		actions = append(actions, b.ab.ItemPickup())
		actions = append(actions, b.ab.ReturnTown())
		running := true
		for running {
			d := game.Status()
			if err := b.hm.HandleHealthAndMana(d); err != nil {
				return err
			}

			for k, act := range actions {
				if !act.Finished(d) {
					err := act.NextStep(d)
					if err != nil {
						fmt.Println(err)
					}
					break
				}
				if len(actions)-1 == k {
					b.logger.Info(fmt.Sprintf("Run %s finished, length: %0.2fs", r.Name(), time.Since(runStart).Seconds()))
					running = false
				}
			}
		}
	}

	return nil
}
