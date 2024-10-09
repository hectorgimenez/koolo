package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/runtype"
	"time"

	"github.com/hectorgimenez/koolo/internal/character"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action"
	botCtx "github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
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

func (b *Bot) Run(ctx context.Context, firstRun bool, runs []runtype.Run) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	gameStartedAt := time.Now()
	b.ctx.SwitchPriority(botCtx.PriorityNormal)
	b.ctx.CurrentGame = &botCtx.CurrentGameHelper{ExpectedArea: b.ctx.Data.PlayerUnit.Area}

	err := b.ctx.GameReader.FetchMapData()
	if err != nil {
		return err
	}

	b.ctx.WaitForGameToLoad()

	action.SwitchToLegacyMode()
	b.ctx.RefreshGameData()

	g.Go(func() error {
		b.ctx.AttachRoutine(botCtx.PriorityBackground)
		ticker := time.NewTicker(10 * time.Millisecond)
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
					if errors.Is(err, health.ErrChicken) || errors.Is(err, health.ErrMercChicken) {
						b.ctx.Logger.Error("Chicken triggered", "error", err)
						// Stop all other goroutines
						b.Stop()

						// Exit game immediately
						exitErr := b.ctx.Manager.ExitGame()
						if exitErr != nil {
							b.ctx.Logger.Error("Failed to exit game", "error", exitErr)
						}
						cancel()
						// Return the chicken error to ensure it's logged properly
						return err
					}
					// For other errors, just cancel and stop
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

				if b.ctx.CharacterCfg.ClassicMode && !b.ctx.Data.LegacyGraphics {
					action.SwitchToLegacyMode()
					b.ctx.RefreshGameData()
				}

				b.ctx.SwitchPriority(botCtx.PriorityHigh)

				// Area correction
				err := b.areaCorrection()
				if err != nil {
					b.ctx.Logger.Error("Area correction failed", "error", err)
				}

				if berserker, ok := b.ctx.Char.(*character.Berserker); !ok || !berserker.IsKillingCouncil() {
					action.ItemPickup(30)
				}
				action.BuffIfRequired()

				_, healingPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.HealingPotion)
				_, manaPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.ManaPotion)
				if (b.ctx.CharacterCfg.BackToTown.NoHpPotions && !healingPotsFound ||
					b.ctx.CharacterCfg.BackToTown.EquipmentBroken && action.RepairRequired() ||
					b.ctx.CharacterCfg.BackToTown.NoMpPotions && !manaPotsFound ||
					b.ctx.CharacterCfg.BackToTown.MercDied && b.ctx.Data.MercHPPercent() <= 0 && b.ctx.CharacterCfg.Character.UseMerc) &&
					!b.ctx.Data.PlayerUnit.Area.IsTown() {
					action.InRunReturnTownRoutine()
				}
				b.ctx.SwitchPriority(botCtx.PriorityNormal)
			}
		}
	})

	g.Go(func() error {
		defer func() {
			cancel()
			b.Stop()
			recover()
		}()

		b.ctx.AttachRoutine(botCtx.PriorityNormal)
		for k, r := range runs {
			b.ctx.CurrentGame.CurrentRun = r // Update the CurrentRun for area correction
			event.Send(event.RunStarted(event.Text(b.ctx.Name, "Starting run"), r.Name()))
			err = action.PreRun(firstRun)
			if err != nil {
				return err
			}

			firstRun = false
			err = r.Run()
			if err != nil {
				return err
			}

			err = action.PostRun(k == len(runs)-1)
			if err != nil {
				return err
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
func (b *Bot) areaCorrection() error {
	// Skip correction if in town
	if b.ctx.Data.PlayerUnit.Area.IsTown() {
		return nil
	}

	if areaAwareRun, ok := b.ctx.CurrentGame.CurrentRun.(run.AreaAwareRun); ok {
		currentArea := b.ctx.Data.PlayerUnit.Area
		if !areaAwareRun.IsAreaPartOfRun(currentArea) {
			b.ctx.Logger.Info("Area mismatch detected", "current", currentArea)
			expectedAreas := areaAwareRun.ExpectedAreas()
			if len(expectedAreas) > 0 {
				err := action.MoveToArea(expectedAreas[0])
				if err != nil {
					b.ctx.Logger.Error("Failed to move to expected area", "error", err)
				}
			}
		}
	}
	return nil
}
