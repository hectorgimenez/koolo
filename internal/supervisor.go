package koolo

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/config"
	"golang.org/x/sync/errgroup"
)

// Supervisor is the main bot entrypoint, it will handle all the parallel processes and ensure everything is up and running
type Supervisor struct {
	cfg           config.Config
	healthManager HealthManager
	bot           Bot
}

func NewSupervisor(cfg config.Config, hm HealthManager, bot Bot) Supervisor {
	return Supervisor{
		cfg:           cfg,
		healthManager: hm,
		bot:           bot,
	}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s Supervisor) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	// Main loop will be inside this, will handle bosses and path traveling
	g.Go(func() error {
		return s.bot.Start(ctx)
	})

	// Will keep our character and mercenary alive, monitoring life and mana
	g.Go(func() error {
		return s.healthManager.Start(ctx)
	})

	return g.Wait()
}
