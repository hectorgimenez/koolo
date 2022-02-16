package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/item"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger *zap.Logger
	cfg    config.Config
	hm     health.Manager
	bm     health.BeltManager
	tm     town.Manager
	char   character.Character
	pickup item.Pickup
}

func NewBot(
	logger *zap.Logger,
	cfg config.Config,
	hm health.Manager,
	bm health.BeltManager,
	tm town.Manager,
	char character.Character,
	pickup item.Pickup,
) Bot {
	return Bot{
		logger: logger,
		cfg:    cfg,
		hm:     hm,
		bm:     bm,
		tm:     tm,
		char:   char,
		pickup: pickup,
	}
}

func (b Bot) RunGame(ctx context.Context, runs []run.Run) map[string]*RunStats {
	defer func() {
		if r := recover(); r == game.NotInGameErr {
		}
	}()

	stats := map[string]*RunStats{}
	totalTime := time.Now()
	for i, r := range runs {
		start := time.Now()
		stats[r.Name()] = &RunStats{}
		if i == 0 {
			b.prepare(true)
		} else {
			b.prepare(false)
		}

		err := r.MoveToStartingPoint()
		if err != nil {
			stats[r.Name()].Errors = 1
			stats[r.Name()].Time = time.Since(start)
			b.logger.Error(fmt.Sprintf("Error moving to start point for %s, let's skip it", r.Name()))
			continue
		}

		ctx, cancel := context.WithCancel(ctx)
		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			return b.hm.Start(ctx)
		})

		g.Go(func() error {
			err = r.TravelToDestination()
			if err != nil {
				b.quitIfErr(totalTime, b.char.ReturnToTown())
				cancel()
				return err
			}

			err = r.Kill()
			if err != nil {
				b.quitIfErr(totalTime, b.char.ReturnToTown())
				cancel()
				return err
			}
			b.logger.Debug(fmt.Sprintf("%s cleared, picking up items...", r.Name()))
			stats[r.Name()].ItemCounter = b.pickup.Pickup()
			stats[r.Name()].Kills = 1

			// Don't return to town on last run, just quit
			if len(runs)-1 != i {
				b.logger.Debug("Item pickup completed, returning to town...")
				b.quitIfErr(totalTime, b.char.ReturnToTown())
			}
			stats[r.Name()].Time = time.Since(start)

			cancel()
			return nil
		})

		err = g.Wait()
		if err != nil {
			stats[r.Name()].Errors = 1
			stats[r.Name()].Time = time.Since(start)
			return stats
		}
	}

	helper.ExitGame()

	return stats
}

func (b Bot) quitIfErr(startTime time.Time, err error) {
	if err != nil {
		helper.ExitGame()
		b.logger.Info(fmt.Sprintf("Game finished with errors. Run time: %0.2f seconds, error: %s", time.Since(startTime).Seconds(), err.Error()))
	}
}
