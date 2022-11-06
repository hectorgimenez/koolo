package koolo

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/memory"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/stats"
	"go.uber.org/zap"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger *zap.Logger
	hm     health.Manager
	ab     action.Builder
	gr     *memory.GameReader
}

func NewBot(
	logger *zap.Logger,
	hm health.Manager,
	ab action.Builder,
	gr *memory.GameReader,
) Bot {
	return Bot{
		logger: logger,
		hm:     hm,
		ab:     ab,
		gr:     gr,
	}
}

func (b *Bot) Run(ctx context.Context, firstRun bool, runs []run.Run) error {
	gameStartedAt := time.Now()

	// TODO: Warmup cache, find a better way to do this shit
	b.logger.Debug("Fetching map data...")
	b.gr.GetData(true)
	b.logger.Debug("Fetch completed", zap.Int64("ms", time.Since(gameStartedAt).Milliseconds()))

	for k, r := range runs {
		stats.StartRun(r.Name())
		runStart := time.Now()
		b.logger.Info(fmt.Sprintf("Running: %s", r.Name()))

		actions := []action.Action{
			b.ab.RecoverCorpse(),
			b.ab.IdentifyAll(firstRun),
			b.ab.Stash(firstRun),
			b.ab.VendorRefill(),
			b.ab.ReviveMerc(),
			b.ab.Repair(),
			b.ab.Heal(),
		}
		firstRun = false

		actions = append(actions, r.BuildActions()...)
		actions = append(actions, b.ab.ItemPickup())

		// Don't return town on last run
		if k != len(runs)-1 {
			actions = append(actions, b.ab.ReturnTown())
		}

		running := true
		for running {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
				d := b.gr.GetData(false)

				if err := b.hm.HandleHealthAndMana(d); err != nil {
					return err
				}
				if err := b.shouldEndCurrentGame(gameStartedAt); err != nil {
					return err
				}

				for k, act := range actions {
					err := act.NextStep(b.logger, d)
					if errors.Is(err, action.ErrNoMoreSteps) {
						if len(actions)-1 == k {
							b.logger.Info(fmt.Sprintf("Run %s finished, length: %0.2fs", r.Name(), time.Since(runStart).Seconds()))
							stats.FinishCurrentRun(stats.EventKill)
							running = false
						}
						continue
					}
					if errors.Is(err, action.ErrWillBeRetried) {
						b.logger.Warn("error occurred, will be retried", zap.Error(err))
						break
					}
					if errors.Is(err, action.ErrCanBeSkipped) {
						stats.Events <- stats.EventWithScreenshot(fmt.Sprintf("error occurred on action that can be skipped, game will continue: %s", err.Error()))
						b.logger.Warn("error occurred on action that can be skipped, game will continue", zap.Error(err))
						act.Skip()
						break
					}
					if err != nil {
						stats.FinishCurrentRun(stats.EventError)
						return err
					}
					break
				}
			}
		}
	}

	return nil
}

func (b *Bot) shouldEndCurrentGame(startedAt time.Time) error {
	if time.Since(startedAt).Seconds() > float64(config.Config.MaxGameLength) {
		return fmt.Errorf(
			"max game length reached, try to exit game: %0.2f",
			time.Since(startedAt).Seconds(),
		)
	}

	return nil
}
