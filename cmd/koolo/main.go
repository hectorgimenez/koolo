package main

import (
	"context"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	cfg, _, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := zapLogger.NewLogger(cfg.Debug, cfg.LogFilePath)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer logger.Sync()

	pf := helper.NewPathFinderV2(logger, cfg)
	bm := health.NewBeltManager(logger, cfg)
	hm := health.NewHealthManager(logger, bm, cfg)
	sm := town.NewShopManager(logger, bm)
	char, err := character.BuildCharacter(cfg)
	if err != nil {
		logger.Fatal("Error creating character", zap.Error(err))
	}

	ab := action.NewBuilder(cfg, logger, pf, sm, bm)
	baseRun := run.NewBaseRun(ab, pf, char)
	runs := []run.Run{run.NewPindleskin(baseRun)}
	//pickup := item.NewPickup(logger, bm, pickit)
	bot := koolo.NewBot(logger, cfg, hm, bm, sm, pf, ab)
	supervisor := koolo.NewSupervisor(logger, cfg, bot)

	ctx := context.Background()
	// TODO: Debug mouse
	go func() {
		if cfg.Debug {
			ticker := time.NewTicker(time.Second * 3)

			for {
				select {
				case <-ticker.C:
					//x, y := robotgo.GetMousePos()
					//logger.Debug(fmt.Sprintf("Display mouse position: X %dpx Y%dpx", x-hid.WindowLeftX, y-hid.WindowTopY))
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	err = supervisor.Start(ctx, runs)
	if err != nil {
		log.Fatalf("Error running Koolo: %s", err.Error())
	}
}
