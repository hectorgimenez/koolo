package koolo

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/container"
	"log/slog"
	"time"

	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/run"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger         *slog.Logger
	hm             health.Manager
	ab             *action.Builder
	container      container.Container
	paused         bool
	supervisorName string
}

func NewBot(
	logger *slog.Logger,
	hm health.Manager,
	ab *action.Builder,
	container container.Container,
	supervisorName string,
) *Bot {
	return &Bot{
		logger:         logger,
		hm:             hm,
		ab:             ab,
		container:      container,
		supervisorName: supervisorName,
	}
}

func (b *Bot) Run(ctx context.Context, firstRun bool, runs []run.Run) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fatal error detected, Koolo will try to exit game and create a new one: %v", r)
		}
	}()

	gameStartedAt := time.Now()
	loadingScreensDetected := 0

	for k, r := range runs {
		event.Send(event.RunStarted(event.Text(b.supervisorName, "Starting run"), r.Name()))
		runStart := time.Now()
		b.logger.Info(fmt.Sprintf("Running: %s", r.Name()))

		actions := b.ab.PreRun(firstRun)
		actions = append(actions, r.BuildActions()...)
		actions = append(actions, b.postRunActions(k, runs)...)

		firstRun = false
		running := true
		loopTime := time.Now()
		var buffAct *action.StepChainAction
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
				if time.Since(loopTime) < time.Millisecond*10 {
					time.Sleep(time.Millisecond*10 - time.Since(loopTime))
				}

				d := b.container.Reader.GetData(false)

				// Skip running stuff if loading screen is present
				if d.OpenMenus.LoadingScreen {
					if loadingScreensDetected == 15 {
						b.logger.Debug("Loading screen detected, waiting until loading screen is gone")
					}
					loadingScreensDetected++
					continue
				}

				if loadingScreensDetected >= 15 {
					b.logger.Debug("Load completed, continuing execution")
				}
				loadingScreensDetected = 0

				if err := b.hm.HandleHealthAndMana(d); err != nil {
					return err
				}

				// Check if game length is exceeded, only if it's not a leveling run
				if r.Name() != run.NameLeveling {
					if err := b.maxGameLengthExceeded(gameStartedAt); err != nil {
						return err
					}
				}

				// TODO: Maybe add some kind of "on every iteration action", something that can be executed/skipped on every iteration
				if b.ab.IsRebuffRequired(d) && (buffAct == nil || buffAct.Steps == nil || buffAct.Steps[len(buffAct.Steps)-1].Status(d, b.container) == step.StatusCompleted) {
					buffAct = b.ab.BuffIfRequired(d)
					actions = append([]action.Action{buffAct}, actions...)
				}

				for k, act := range actions {
					err := act.NextStep(d, b.container)
					loopTime = time.Now()
					if errors.Is(err, action.ErrNoMoreSteps) {
						if len(actions)-1 == k {
							b.logger.Info(fmt.Sprintf("Run %s finished, length: %0.2fs", r.Name(), time.Since(runStart).Seconds()))
							event.Send(event.RunFinished(event.Text(b.supervisorName, "Finished run"), r.Name(), event.FinishedOK))
							running = false
						}
						continue
					}
					if errors.Is(err, action.ErrWillBeRetried) {
						b.logger.Warn("error occurred, will be retried", slog.Any("error", err))
						break
					}
					if errors.Is(err, action.ErrCanBeSkipped) {
						event.Send(event.RunFinished(event.WithScreenshot(b.supervisorName, err.Error(), b.container.Reader.Screenshot()), r.Name(), event.FinishedError))
						b.logger.Warn("error occurred on action that can be skipped, game will continue", slog.Any("error", err))
						act.Skip()
						break
					}
					if errors.Is(err, action.ErrLogAndContinue) {
						b.logger.Warn(err.Error())
						break
					}
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return nil
}

func (b *Bot) maxGameLengthExceeded(startedAt time.Time) error {
	if time.Since(startedAt).Seconds() > float64(config.Characters[b.supervisorName].MaxGameLength) {
		return fmt.Errorf(
			"max game length reached, try to exit game: %0.2f",
			time.Since(startedAt).Seconds(),
		)
	}

	return nil
}

func (b *Bot) postRunActions(currentRun int, runs []run.Run) []action.Action {
	if config.Characters[b.supervisorName].Companion.Enabled && !config.Characters[b.supervisorName].Companion.Leader {
		return []action.Action{}
	}

	actions := []action.Action{
		b.ab.ClearAreaAroundPlayer(5),
		b.ab.ItemPickup(true, -1),
	}

	// Don't return town on last run
	if currentRun != len(runs)-1 {
		if config.Characters[b.supervisorName].Game.ClearTPArea {
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
