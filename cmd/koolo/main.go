package main

import (
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
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

	tf, err := koolo.NewTemplateFinder(logger, "assets/templates")
	d, err := koolo.NewDisplay(cfg.Display, logger)

	hm := koolo.NewHealthManager(d, tf)
	bot := koolo.NewBot()
	supervisor := koolo.NewSupervisor(cfg, hm, bot)

	ctx := context.Background()
	// TODO: Debug mouse
	go func() {
		if cfg.Debug {
			ticker := time.NewTicker(time.Second * 3)

			for {
				select {
				case <-ticker.C:
					x, y := robotgo.GetMousePos()
					relativeX := x - d.OffsetLeft
					relativeY := y - d.OffsetTop
					logger.Debug(fmt.Sprintf("Window mouse position: X %dpx Y%dpx", relativeX, relativeY))
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
