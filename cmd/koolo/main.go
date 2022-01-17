package main

import (
	"context"
	"log"

	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	supervisor := koolo.NewSupervisor(cfg)

	ctx := context.Background()
	err = supervisor.Start(ctx)
	_, err = koolo.NewTemplateFinder("assets/templates")
	if err != nil {
		log.Fatalf("Error running Koolo: %s", err.Error())
	}
}
