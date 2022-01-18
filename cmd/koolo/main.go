package main

import (
	"context"
	zapLogger "github.com/hectorgimenez/koolo/cmd/koolo/log"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
	"log"
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
	err = supervisor.Start(ctx)

	if err != nil {
		log.Fatalf("Error running Koolo: %s", err.Error())
	}
}
