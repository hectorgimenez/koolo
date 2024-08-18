package bot

import (
	"context"
	"time"

	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/v2/action"
	botCtx "github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/run"
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
	g, ctx := errgroup.WithContext(ctx)

	// Let's make sure we have updated game data before we start the runs
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
				return nil
			case <-ticker.C:
				err := b.ctx.HealthManager.HandleHealthAndMana()
				if err != nil {
					cancel()
					b.Stop()
					return err
				}
			}
		}
	})

	// High priority loop, this will interrupt (pause) low priority loop
	g.Go(func() error {
		b.ctx.AttachRoutine(botCtx.PriorityHigh)
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if len(action.GetItemsToPickup(30)) > 0 {
					b.ctx.ExecutionPriority = botCtx.PriorityHigh
					_ = action.ItemPickup(30)
					b.ctx.ExecutionPriority = botCtx.PriorityNormal
				}
			}
		}
	})

	// Low priority loop, this will keep executing main run scripts
	g.Go(func() error {
		b.ctx.AttachRoutine(botCtx.PriorityNormal)
		for _, r := range runs {
			event.Send(event.RunStarted(event.Text(b.ctx.Name, "Starting run"), r.Name()))
			err := action.PreRun(firstRun)
			if err != nil {
				return err
			}

			err = r.Run()
			if err != nil {
				return err
			}
			err = action.PostRun(false)
			if err != nil {
				return err
			}
			err = action.ItemPickup(30)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return g.Wait()
}

func (b *Bot) Stop() {
	b.ctx.Cancel()
}
