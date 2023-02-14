package koolo

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/event/stat"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/memory"
	"github.com/hectorgimenez/koolo/internal/run"
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
		stat.StartRun(r.Name())
		runStart := time.Now()
		b.logger.Info(fmt.Sprintf("Running: %s", r.Name()))

		actions := []action.Action{
			b.ab.RecoverCorpse(),
			b.ab.IdentifyAll(firstRun),
			b.ab.Stash(firstRun),
			b.ab.VendorRefill(),
			b.ab.Heal(),
			b.ab.ReviveMerc(),
			b.ab.Repair(),
		}
		firstRun = false

		actions = append(actions, r.BuildActions()...)
		actions = append(actions, b.ab.ClearAreaAroundPlayer(5))
		actions = append(actions, b.ab.ItemPickup(true, -1))

		// Don't return town on last run
		if k != len(runs)-1 {
			if config.Config.Game.ClearTPArea {
				actions = append(actions, b.ab.ClearAreaAroundPlayer(5))
				actions = append(actions, b.ab.ItemPickup(false, -1))
			}
			actions = append(actions, b.ab.ReturnTown())
		}

		running := true
		loopTime := time.Now()
		for running {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
				// Throttle loop a bit, don't need to waste CPU
				if time.Since(loopTime) < time.Millisecond*30 {
					time.Sleep(time.Millisecond*30 - time.Since(loopTime))
				}

				d := b.gr.GetData(false)

				// Skip running stuff if loading screen is present
				if d.OpenMenus.LoadingScreen {
					continue
				}

				if err := b.hm.HandleHealthAndMana(d); err != nil {
					return err
				}
				if err := b.shouldEndCurrentGame(gameStartedAt); err != nil {
					return err
				}

				for k, act := range actions {
					err := act.NextStep(b.logger, d)
					loopTime = time.Now()
					if errors.Is(err, action.ErrNoMoreSteps) {
						if len(actions)-1 == k {
							b.logger.Info(fmt.Sprintf("Run %s finished, length: %0.2fs", r.Name(), time.Since(runStart).Seconds()))
							stat.FinishCurrentRun(event.Kill)
							running = false
						}
						continue
					}
					if errors.Is(err, action.ErrWillBeRetried) {
						b.logger.Warn("error occurred, will be retried", zap.Error(err))
						break
					}
					if errors.Is(err, action.ErrCanBeSkipped) {
						event.Events <- event.WithScreenshot(fmt.Sprintf("error occurred on action that can be skipped, game will continue: %s", err.Error()))
						b.logger.Warn("error occurred on action that can be skipped, game will continue", zap.Error(err))
						act.Skip()
						break
					}
					if err != nil {
						stat.FinishCurrentRun(event.Error)
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
