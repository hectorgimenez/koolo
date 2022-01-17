package main

import (
	"context"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	logger, err := NewLogger(cfg.Debug, cfg.LogFilePath)
	if err != nil {
		log.Fatalf("Error starting logger: %s", err.Error())
	}
	defer logger.Sync()

	supervisor := koolo.NewSupervisor(cfg)

	ctx := context.Background()
	err = supervisor.Start(ctx)
	_, err = koolo.NewTemplateFinder(logger, "assets/templates")
	if err != nil {
		log.Fatalf("Error running Koolo: %s", err.Error())
	}
}
