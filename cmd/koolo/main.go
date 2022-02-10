package main

import (
	"context"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/item"
	"github.com/hectorgimenez/koolo/internal/mapassist"
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
	mapAssistApi := mapassist.NewAPIClient(cfg.MapAssist.HostName)
	bm := health.NewBeltManager(logger, cfg, mapAssistApi)
	hm := health.NewHealthManager(logger, mapAssistApi, chEvents, bm, cfg)
	pf := helper.NewPathFinder(logger, mapAssistApi, cfg)
	sm := town.NewShopManager(logger, mapAssistApi, bm)
	tm := town.NewTownManager(cfg, mapAssistApi, pf, sm)
	char, err := character.BuildCharacter(mapAssistApi, cfg)
	if err != nil {
		logger.Fatal("Error creating character", zap.Error(err))
	}
	baseRun := run.NewBaseRun(mapAssistApi, pf, char)
	runs := []run.Run{run.NewPindleskin(baseRun)}
	pickup := item.NewPickup(logger, mapAssistApi, bm, pf, pickit)

	bot := game.NewBot(logger, cfg, bm, tm, mapAssistApi, char, runs, pickup)
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

	err = supervisor.Start(ctx)
	if err != nil {
		log.Fatalf("Error running Koolo: %s", err.Error())
	}
}
