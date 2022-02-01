package main

import (
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/mapassist"
	"github.com/hectorgimenez/koolo/internal/town"
	"log"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := zapLogger.NewLogger(cfg.Debug, cfg.LogFilePath)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer logger.Sync()

	chActions := make(chan action.Action, 10)
	chEvents := make(chan event.Event, 10)
	ah := action.NewHandler(chActions)
	mapAssistApi := mapassist.NewAPIClient(cfg.MapAssist.HostName)
	bm := health.NewBeltManager(logger, cfg, mapAssistApi, chActions)
	hm := health.NewHealthManager(logger, mapAssistApi, chActions, chEvents, bm, cfg)
	pf := helper.NewPathFinder(logger, mapAssistApi)
	tm := town.NewTownManager(mapAssistApi, pf, chActions)
	bot := game.NewBot(logger, cfg, bm, tm, mapAssistApi, chActions)
	supervisor := koolo.NewSupervisor(logger, cfg, ah, hm, bot)

	ctx := context.Background()
	// TODO: Debug mouse
	go func() {
		if cfg.Debug {
			ticker := time.NewTicker(time.Second * 3)

			for {
				select {
				case <-ticker.C:
					x, y := robotgo.GetMousePos()
					logger.Debug(fmt.Sprintf("Display mouse position: X %dpx Y%dpx", x-hid.WindowLeftX, y-hid.WindowTopY))
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
