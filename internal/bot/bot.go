package bot

import (
	"context"
	"fmt"
	"time"

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

func (b *Bot) Run(ctx context.Context, firstRun bool, runs []run.Run) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	gameStartedAt := time.Now()
	b.ctx.SwitchPriority(botCtx.PriorityNormal) // Restore priority to normal, in case it was stopped in previous game

	// Let's make sure we have updated game data before we start the runs
	err := b.ctx.GameReader.FetchMapData()
	if err != nil {
		return err
	}

	b.ctx.RefreshGameData()

	// This routine is in charge of refreshing the game data and handling cancellation, will work in parallel with any other execution
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
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if b.ctx.ExecutionPriority == botCtx.PriorityPause {
					continue
				}

				b.ctx.SwitchPriority(botCtx.PriorityHigh)
				action.SwitchToLegacyMode()
				action.ItemPickup(30)
				action.BuffIfRequired()

				_, healingPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.HealingPotion)
				_, manaPotsFound := b.ctx.Data.Inventory.Belt.GetFirstPotion(data.ManaPotion)
				// Check if we need to go back to town (no pots or merc died)
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

	// Low priority loop, this will keep executing main run scripts
	g.Go(func() error {
		defer func() {
			cancel()
			b.Stop()
			recover()
		}()

		b.ctx.AttachRoutine(botCtx.PriorityNormal)
		for k, r := range runs {
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
