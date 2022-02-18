package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger *zap.Logger
	cfg    config.Config
	hm     health.Manager
	bm     health.BeltManager
	sm     town.ShopManager
	pf     helper.PathFinderV2
	ab     action.Builder
}

func NewBot(
	logger *zap.Logger,
	cfg config.Config,
	hm health.Manager,
	bm health.BeltManager,
	sm town.ShopManager,
	pf helper.PathFinderV2,
	ab action.Builder,
) Bot {
	return Bot{
		logger: logger,
		cfg:    cfg,
		hm:     hm,
		bm:     bm,
		sm:     sm,
		pf:     pf,
		ab:     ab,
	}
}

func (b *Bot) Run(ctx context.Context, runs []run.Run) error {
	for _, r := range runs {
		runStart := time.Now()
		b.logger.Debug(fmt.Sprintf("Running: %s", r.Name()))

		actions := []action.Action{
			b.ab.RecoverCorpse(),
			b.ab.VendorRefill(),
			b.ab.ReviveMerc(),
			b.ab.Repair(),
			b.ab.Heal(),
			b.ab.Stash(),
		}

		actions = append(actions, r.BuildActions(game.Status())...)
		running := true
		for running {
			d := game.Status()
			b.hm.HandleHealthAndMana(d)

			for k, act := range actions {
				if !act.Finished(d) {
					err := act.NextStep(d)
					if err != nil {
						fmt.Println(err)
					}
					break
				}
				if len(actions)-1 == k {
					b.logger.Info(fmt.Sprintf("Run %s finished, length: %0.2fs", r.Name(), time.Since(runStart).Seconds()))
					running = false
				}
			}
		}
	}

	return nil
}

//func (b Bot) RunGame(ctx context.Context, runs []run.Run) map[string]*RunStats {
//	defer func() {
//		if r := recover(); r == game.NotInGameErr {
//		}
//	}()
//
//	stats := map[string]*RunStats{}
//	totalTime := time.Now()
//	for i, r := range runs {
//		start := time.Now()
//		stats[r.Name()] = &RunStats{}
//		if i == 0 {
//			b.prepare(true)
//		} else {
//			b.prepare(false)
//		}
//
//		err := r.MoveToStartingPoint()
//		if err != nil {
//			stats[r.Name()].Errors = 1
//			stats[r.Name()].Time = time.Since(start)
//			b.logger.Error(fmt.Sprintf("Error moving to start point for %s, let's skip it", r.Name()))
//			continue
//		}
//
//		ctx, cancel := context.WithCancel(ctx)
//		g, ctx := errgroup.WithContext(ctx)
//		g.Go(func() error {
//			return b.hm.Start(ctx)
//		})
//
//		g.Go(func() error {
//			err = r.TravelToDestination()
//			if err != nil {
//				b.quitIfErr(totalTime, b.char.ReturnToTown())
//				cancel()
//				return err
//			}
//
//			err = r.Kill()
//			if err != nil {
//				b.quitIfErr(totalTime, b.char.ReturnToTown())
//				cancel()
//				return err
//			}
//			b.logger.Debug(fmt.Sprintf("%s cleared, picking up items...", r.Name()))
//			stats[r.Name()].ItemCounter = b.pickup.Pickup()
//			stats[r.Name()].Kills = 1
//			b.logger.Debug("Item pickup completed")
//
//			// Don't return to town on last run, just quit
//			if len(runs)-1 != i {
//				b.logger.Debug("Returning to town...")
//				b.quitIfErr(totalTime, b.char.ReturnToTown())
//			}
//			stats[r.Name()].Time = time.Since(start)
//
//			cancel()
//			return nil
//		})
//
//		err = g.Wait()
//		if err != nil {
//			stats[r.Name()].Errors = 1
//			stats[r.Name()].Time = time.Since(start)
//			return stats
//		}
//	}
//
//	b.logger.Info(fmt.Sprintf("Game finished successfully. Run time: %0.2f seconds", time.Since(totalTime).Seconds()))
//	helper.ExitGame()
//
//	return stats
//}
//
//func (b Bot) quitIfErr(startTime time.Time, err error) {
//	if err != nil {
//		helper.ExitGame()
//		b.logger.Info(fmt.Sprintf("Game finished with errors. Run time: %0.2f seconds, error: %s", time.Since(startTime).Seconds(), err.Error()))
//	}
//}
