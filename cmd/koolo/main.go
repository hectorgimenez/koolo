package main

import (
	"context"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/item"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	cfg, pickit, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := zapLogger.NewLogger(cfg.Debug, cfg.LogFilePath)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer logger.Sync()

	chEvents := make(chan event.Event, 0)
	bm := health.NewBeltManager(logger, cfg)
	hm := health.NewHealthManager(logger, chEvents, bm, cfg)
	pf := helper.NewPathFinder(logger, cfg)
	sm := town.NewShopManager(logger, bm)
	tm := town.NewTownManager(cfg, pf, sm)
	char, err := character.BuildCharacter(cfg, pf)
	if err != nil {
		logger.Fatal("Error creating character", zap.Error(err))
	}
	baseRun := run.NewBaseRun(pf, char, tm)
	runs := []run.Run{run.NewAndariel(baseRun), run.NewSummoner(baseRun), run.NewMephisto(baseRun), run.NewPindleskin(baseRun)}
	pickup := item.NewPickup(logger, bm, pf, pickit)

	bot := koolo.NewBot(logger, cfg, hm, bm, tm, char, pickup)
	supervisor := koolo.NewSupervisor(logger, cfg, hm, bot)

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
