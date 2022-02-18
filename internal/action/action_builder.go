package action

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
)

type Builder struct {
	cfg    config.Config
	logger *zap.Logger
	pf     helper.PathFinderV2
	sm     town.ShopManager
	bm     health.BeltManager
}

func NewBuilder(cfg config.Config, logger *zap.Logger, pf helper.PathFinderV2, sm town.ShopManager, bm health.BeltManager) Builder {
	return Builder{
		cfg:    cfg,
		logger: logger,
		pf:     pf,
		sm:     sm,
		bm:     bm,
	}
}
