package koolo

import (
	"context"
	"github.com/hectorgimenez/koolo/internal/config"
)

// Supervisor is the main bot entrypoint, it will handle all the parallel processes and ensure everything is up and running
type Supervisor struct {
	cfg config.Config
}

func NewSupervisor(cfg config.Config) Supervisor {
	return Supervisor{cfg: cfg}
}

// Start will stay running during the application lifecycle, it will orchestrate all the required bot pieces
func (s Supervisor) Start(ctx context.Context) error {
	bot := NewBot()
	healthManager := NewHealthManager()

	bot.Start(ctx)
	healthManager.Start(ctx)

	return nil
}
