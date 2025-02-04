package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action"
	botCtx "github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/run"
	"golang.org/x/sync/errgroup"
)

type Bot struct {
	ctx *botCtx.Context
}

func NewBot(ctx *botCtx.Context) *Bot {
	return &Bot{
		ctx: ctx,
	}
}
func (b *Bot) Run(ctx context.Context, firstRun bool, runs []run.Run) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	gameStartedAt := time.Now()
	b.ctx.SwitchPriority(botCtx.PriorityNormal) // Restore priority to normal, in case it was stopped in previous game
	b.ctx.CurrentGame = botCtx.NewGameHelper()  // Reset current game helper structure

	err := b.ctx.GameReader.FetchMapData()
	if err != nil {
		return err
	}

	// Let's make sure we have updated game data also fully loaded before performing anything
	b.ctx.WaitForGameToLoad()
	// Switch to legacy mode if configured
	action.SwitchToLegacyMode()
	b.ctx.RefreshGameData()

	// This routine is in charge of refreshing the game data and handling cancellation, will work in parallel with any other execution
	g.Go(func() error {
		b.ctx.AttachRoutine(botCtx.PriorityBackground)
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				cancel()
				b.Stop()
				return nil
			case <-ticker.C:
				if b.ctx.ExecutionPriority == botCtx.PriorityPause {
					continue
				}
				b.ctx.RefreshGameData()
			}
		}
	})

	// This routine is in charge of handling the health/chicken of the bot, will work in parallel with any other execution
	g.Go(func() error {
		b.ctx.AttachRoutine(botCtx.PriorityBackground)
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				b.Stop()
				return nil
			case <-ticker.C:
				if b.ctx.ExecutionPriority == botCtx.PriorityPause {
					continue
				}
				err = b.ctx.HealthManager.HandleHealthAndMana()
				if err != nil {
					cancel()
					b.Stop()
					return err
				}
				if time.Since(gameStartedAt).Seconds() > float64(b.ctx.CharacterCfg.MaxGameLength) {
					cancel()
					b.Stop()
					return fmt.Errorf(
						"max game length reached, try to exit game: %0.2f",
						time.Since(gameStartedAt).Seconds(),
					)
				}
			}
		}
	})
	// High priority loop, this will interrupt (pause) low priority loop
	g.Go(func() error {
		defer func() {
			cancel()
			b.Stop()
			recover()
		}()

		b.ctx.AttachRoutine(botCtx.PriorityHigh)
		ticker := time.NewTicker(time.Millisecond * 100)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if b.ctx.ExecutionPriority == botCtx.PriorityPause {
					continue
				}

				// Sometimes when we switch areas, monsters are not loaded yet, and we don't properly detect the Merc
				// let's add some small delay (just few ms) when this happens, and recheck the merc status
				if b.ctx.CharacterCfg.BackToTown.MercDied && b.ctx.Data.MercHPPercent() <= 0 && b.ctx.CharacterCfg.Character.UseMerc {
					time.Sleep(200 * time.Millisecond)
				}

				// extra RefreshGameData not needed for Legacygraphics/Portraits since Background loop will automatically refresh after 100ms
				if b.ctx.CharacterCfg.ClassicMode && !b.ctx.Data.LegacyGraphics {
					// Toggle Legacy if enabled
					action.SwitchToLegacyMode()
					time.Sleep(150 * time.Millisecond)
				}
				// Hide merc/other players portraits if enabled
				if b.ctx.CharacterCfg.HidePortraits && b.ctx.Data.OpenMenus.PortraitsShown {
					action.HidePortraits()
					time.Sleep(150 * time.Millisecond)
				}
				// Close chat if somehow was opened (prevention)
				if b.ctx.Data.OpenMenus.ChatOpen {
					b.ctx.HID.PressKey(b.ctx.Data.KeyBindings.Chat.Key1[0])
					time.Sleep(150 * time.Millisecond)
				}
				b.ctx.SwitchPriority(botCtx.PriorityHigh)

				// Area correction (only check if enabled)
				if b.ctx.CurrentGame.AreaCorrection.Enabled {
					if err = action.AreaCorrection(); err != nil {
						b.ctx.Logger.Warn("Area correction failed", "error", err)
					}
				}

				// Perform item pickup if enabled
				if b.ctx.CurrentGame.PickupItems {
					action.ItemPickup(30)
				}
				action.BuffIfRequired()

				_, healingPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.HealingPotion)
				_, manaPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.ManaPotion)

				// Check if we need to go back to town (no pots or merc died)
				if (b.ctx.CharacterCfg.BackToTown.NoHpPotions && !healingPotsFound ||
					b.ctx.CharacterCfg.BackToTown.EquipmentBroken && action.RepairRequired() ||
					b.ctx.CharacterCfg.BackToTown.NoMpPotions && !manaPotsFound ||
					b.ctx.CharacterCfg.BackToTown.MercDied && b.ctx.Data.MercHPPercent() <= 0 && b.ctx.CharacterCfg.Character.UseMerc) &&
					!b.ctx.Data.PlayerUnit.Area.IsTown() {

					// Log the exact reason for going back to town
					var reason string
					if b.ctx.CharacterCfg.BackToTown.NoHpPotions && !healingPotsFound {
						reason = "No healing potions found"
					} else if b.ctx.CharacterCfg.BackToTown.EquipmentBroken && action.RepairRequired() {
						reason = "Equipment broken"
					} else if b.ctx.CharacterCfg.BackToTown.NoMpPotions && !manaPotsFound {
						reason = "No mana potions found"
					} else if b.ctx.CharacterCfg.BackToTown.MercDied && b.ctx.Data.MercHPPercent() <= 0 && b.ctx.CharacterCfg.Character.UseMerc {
						reason = "Mercenary is dead"
					}

					b.ctx.Logger.Info("Going back to town", "reason", reason)

					// Try to return to town up to 3 times
					maxRetries := 3
					var lastError error
					for attempt := 0; attempt < maxRetries; attempt++ {
						if err := action.InRunReturnTownRoutine(); err != nil {
							lastError = err
							b.ctx.Logger.Warn("Failed to return to town", "error", err, "attempt", attempt+1)
							// Wait a bit before retrying
							time.Sleep(500 * time.Millisecond)
							continue
						}
						lastError = nil
						break
					}

					// If we still failed after retries, log it but continue running
					if lastError != nil {
						b.ctx.Logger.Error("Failed to return to town after all retries", "error", lastError)
					}
				}
				b.ctx.SwitchPriority(botCtx.PriorityNormal)
			}
		}
	})

	// Low priority loop, this will keep executing main run scripts
	g.Go(func() error {
		defer func() {
			cancel()
			b.Stop()
			recover()
		}()

		b.ctx.AttachRoutine(botCtx.PriorityNormal)
		for _, r := range runs {
			select {
			case <-ctx.Done():
				return nil
			default:
				event.Send(event.RunStarted(event.Text(b.ctx.Name, fmt.Sprintf("Starting run: %s", r.Name())), r.Name()))
				err = action.PreRun(firstRun)
				if err != nil {
					return err
				}

				firstRun = false
				err = r.Run()

				var runFinishReason event.FinishReason
				if err != nil {
					switch {
					case errors.Is(err, health.ErrChicken):
						runFinishReason = event.FinishedChicken
					case errors.Is(err, health.ErrMercChicken):
						runFinishReason = event.FinishedMercChicken
					case errors.Is(err, health.ErrDied):
						runFinishReason = event.FinishedDied
					default:
						runFinishReason = event.FinishedError
					}
				} else {
					runFinishReason = event.FinishedOK
				}

				event.Send(event.RunFinished(event.Text(b.ctx.Name, fmt.Sprintf("Finished run: %s", r.Name())), r.Name(), runFinishReason))

				if err != nil {
					return err
				}

				err = action.PostRun(r == runs[len(runs)-1])
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	return g.Wait()
}

func (b *Bot) Stop() {
	b.ctx.SwitchPriority(botCtx.PriorityStop)
	b.ctx.Detach()
}
