package koolo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/event/stat"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/run"
	"go.uber.org/zap"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger *zap.Logger
	hm     health.Manager
	ab     action.Builder
	gr     *reader.GameReader
	paused bool
}

func NewBot(
	logger *zap.Logger,
	hm health.Manager,
	ab action.Builder,
	gr *reader.GameReader,
) *Bot {
	return &Bot{
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

	loadingScreensDetected := 0
	for k, r := range runs {
		stat.StartRun(r.Name())
		runStart := time.Now()
		b.logger.Info(fmt.Sprintf("Running: %s", r.Name()))

		actions := b.preRunActions(firstRun)
		actions = append(actions, r.BuildActions()...)
		actions = append(actions, b.postRunActions(k, runs)...)

		firstRun = false
		running := true
		loopTime := time.Now()
		for running {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
				if b.paused {
					time.Sleep(time.Second)
					continue
				}

				// Throttle loop a bit, don't need to waste CPU
				if time.Since(loopTime) < time.Millisecond*30 {
					time.Sleep(time.Millisecond*30 - time.Since(loopTime))
				}

				d := b.gr.GetData(false)

				// Skip running stuff if loading screen is present
				if d.OpenMenus.LoadingScreen {
					if loadingScreensDetected == 10 {
						b.logger.Debug("Loading screen detected, waiting until loading screen is gone")
					}
					loadingScreensDetected++
					continue
				} else {
					if loadingScreensDetected >= 10 {
						b.logger.Debug("Load completed, continuing execution")
					}
					loadingScreensDetected = 0
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
					if errors.Is(err, action.ErrLogAndContinue) {
						b.logger.Warn(err.Error())
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

func (b *Bot) preRunActions(firstRun bool) []action.Action {
	if config.Config.Companion.Enabled && !config.Config.Companion.Leader {
		return []action.Action{
			b.ab.RecoverCorpse(),
			b.ab.Heal(),
		}
	}

	return []action.Action{
		b.ab.EnsureEmptyHand(),
		b.ab.EnsureStatPoints(),
		b.ab.EnsureSkillPoints(),
		b.ab.RecoverCorpse(),
		b.ab.IdentifyAll(firstRun),
		b.ab.Stash(firstRun),
		b.ab.VendorRefill(),
		b.ab.Heal(),
		b.ab.ReviveMerc(),
		b.ab.Repair(),
	}
}

func (b *Bot) postRunActions(currentRun int, runs []run.Run) []action.Action {
	if config.Config.Companion.Enabled && !config.Config.Companion.Leader {
		return []action.Action{}
	}

	actions := []action.Action{
		b.ab.ClearAreaAroundPlayer(5),
		b.ab.ItemPickup(true, -1),
	}

	// Don't return town on last run
	if currentRun != len(runs)-1 {
		if config.Config.Game.ClearTPArea {
			actions = append(actions, b.ab.ClearAreaAroundPlayer(5))
			actions = append(actions, b.ab.ItemPickup(false, -1))
		}
		actions = append(actions, b.ab.ReturnTown())
	}

	return actions
}

func (b *Bot) TogglePause() {
	if b.paused {
		b.logger.Info("Resuming...")
		b.paused = false
	} else {
		b.logger.Info("Pausing...")
		b.paused = true
	}
}
