package koolo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/run"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger          *slog.Logger
	hm              *health.Manager
	ab              *action.Builder
	c               container.Container
	paused          bool
	pauseRequested  bool
	resumeRequested bool
	supervisorName  string
}

func NewBot(
	logger *slog.Logger,
	hm *health.Manager,
	ab *action.Builder,
	container container.Container,
	supervisorName string,
) *Bot {
	return &Bot{
		logger:         logger,
		hm:             hm,
		ab:             ab,
		c:              container,
		supervisorName: supervisorName,
	}
}

func (b *Bot) Run(ctx context.Context, firstRun bool, runs []run.Run) (err error) {
	companionTPRequestedAt := time.Time{}
	companionTPRequested := false
	companionLeftGame := false
	if b.c.CharacterCfg.Companion.Enabled && b.c.CharacterCfg.Companion.Leader {
		b.c.EventListener.Register(func(ctx context.Context, e event.Event) error {
			switch evt := e.(type) {
			case event.CompanionRequestedTPEvent:
				if time.Since(companionTPRequestedAt) > time.Second*5 {
					companionTPRequestedAt = time.Now()
					companionTPRequested = true
				}
			case event.GameFinishedEvent:
				cmp := config.Characters[evt.Supervisor()].Companion
				if cmp.Enabled && !cmp.Leader {
					companionLeftGame = true
				}
			}

			return nil
		})
	}

	gameStartedAt := time.Now()
	loadingScreensDetected := 0

	actions := b.ab.NewGameHook()

	for k, r := range runs {

		if config.Koolo.Discord.EnableNewRunMessages {
			event.Send(event.RunStarted(event.Text(b.supervisorName, "Starting run"), r.Name()))
		} else {
			event.Send(event.RunStarted(event.Text(b.supervisorName, ""), r.Name()))
		}
		runStart := time.Now()
		b.logger.Info(fmt.Sprintf("Running: %s", r.Name()))

		actions = slices.Concat(actions,
			b.ab.PreRunHook(firstRun),
			r.BuildActions(),
			b.ab.PostRunHook(k == len(runs)-1),
		)
		eachLoopActions := make([]action.Action, 0)

		firstRun = false
		running := true
		loopTime := time.Now()
		for running {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
				if b.resumeRequested {
					b.paused = false
					b.resumeRequested = false
					b.pauseRequested = false
					b.logger.Info("Resuming...")
					b.c.Injector.Load()
					event.Send(event.GamePaused(event.Text(b.supervisorName, "Game resumed"), false))
				}
				if b.pauseRequested {
					b.paused = true
					b.resumeRequested = false
					b.pauseRequested = false
					b.logger.Info("Pausing...")
					b.c.Injector.RestoreMemory()
					event.Send(event.GamePaused(event.Text(b.supervisorName, "Game paused"), true))
					b.paused = true
				}

				if b.paused {
					time.Sleep(time.Second)
					continue
				}

				// Throttle loop a bit, don't need to waste CPU
				if time.Since(loopTime) < time.Millisecond*10 {
					time.Sleep(time.Millisecond*10 - time.Since(loopTime))
				}

				d := b.c.Reader.GetData(false)

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

				// By this point we're ingame

				// Check if we have all keybindings
				missingBindings := b.ab.CheckKeyBindings(d)
				if len(missingBindings) > 0 {
					var str = "Missing skill bindings for skills:"
					for _, id := range missingBindings {
						str += "\n" + skill.SkillNames[id]
					}
					str += "\nPlease bind the skills to a key. Pausing bot..."

					// Display the message box
					helper.ShowDialog(b.supervisorName+" skill bindings missing", str)

					// Pause the bot
					b.pauseRequested = true
					continue
				}

				if err := b.hm.HandleHealthAndMana(d); err != nil {
					return err
				}

				// Check if game length is exceeded, only if it's not a leveling run
				if r.Name() != string(config.LevelingRun) {
					if err := b.maxGameLengthExceeded(gameStartedAt); err != nil {
						return err
					}
				}
				// Some hacky stuff for companion mode, ideally should be encapsulated everything together in a different place
				if d.CharacterCfg.Companion.Enabled {
					if companionTPRequested && r.Name() == string(config.LevelingRun) {
						companionTPRequested = false
						actions = append([]action.Action{b.ab.OpenTPIfLeader()}, actions...)
					}
					if companionLeftGame {
						event.Send(event.RunFinished(event.WithScreenshot(b.supervisorName, "Companion left game", b.c.Reader.Screenshot()), r.Name(), event.FinishedError))
						return errors.New("companion left game")
					}
					_, leaderFound := d.Roster.FindByName(d.CharacterCfg.Companion.LeaderName)
					if !leaderFound {
						event.Send(event.RunFinished(event.WithScreenshot(b.supervisorName, "Leader left game", b.c.Reader.Screenshot()), r.Name(), event.FinishedError))
						return errors.New("leader left game")
					}
				}

				if len(eachLoopActions) == 0 || (reflect.ValueOf(eachLoopActions[len(eachLoopActions)-1]).IsNil() || eachLoopActions[len(eachLoopActions)-1].IsFinished()) {
					eachLoopActions = b.ab.EachLoopHook(d)
					if len(eachLoopActions) > 0 {
						actions = append(eachLoopActions, actions...)
					}
				}

				for k, act := range actions {
					// Ensure we're not trying to access a nil action
					if act == nil {
						continue
					}
					err := act.NextStep(d, b.c)
					loopTime = time.Now()
					if errors.Is(err, action.ErrNoMoreSteps) {
						if len(actions)-1 == k {
							b.logger.Info(fmt.Sprintf("Run %s finished, length: %0.2fs", r.Name(), time.Since(runStart).Seconds()))
							if config.Koolo.Discord.EnableRunFinishMessages {
								event.Send(event.RunFinished(event.Text(b.supervisorName, "Finished run"), r.Name(), event.FinishedOK))
							} else {
								event.Send(event.RunFinished(event.Text(b.supervisorName, ""), r.Name(), event.FinishedOK))
							}
							running = false
						}
						continue
					}
					if errors.Is(err, action.ErrWillBeRetried) {
						b.logger.Warn("error occurred, will be retried", slog.Any("error", err))
						break
					}
					if errors.Is(err, action.ErrCanBeSkipped) {
						event.Send(event.RunFinished(event.WithScreenshot(b.supervisorName, err.Error(), b.c.Reader.Screenshot()), r.Name(), event.FinishedError))
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
	// Check if config or Characters map is nil
	if config.Characters == nil {
		helper.Sleep(1000) // Wait a second a re-check to avoid configuration saved or reloads
		if config.Characters == nil {
			return fmt.Errorf("configuration is not initialized")
		}
	}

	// Check if specific supervisor config is nil
	characterConfig, exists := config.Characters[b.supervisorName]
	if !exists || characterConfig == nil {
		helper.Sleep(1000) // Wait a second and re-check to avoid configuration saves or reloads resulting in errors
		characterConfig, exists := config.Characters[b.supervisorName]
		if !exists || characterConfig == nil {
			return fmt.Errorf("character configuration for %s not found or is nil", b.supervisorName)
		}
	}

	if time.Since(startedAt).Seconds() > float64(config.Characters[b.supervisorName].MaxGameLength) {
		return fmt.Errorf(
			"max game length reached, try to exit game: %0.2f",
			time.Since(startedAt).Seconds(),
		)
	}

	return nil
}

func (b *Bot) TogglePause() {
	if b.paused {
		b.resumeRequested = true
	} else {
		b.pauseRequested = true
	}
}
