package action

import (
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
)

type Builder struct {
	logger *zap.Logger
	sm     town.ShopManager
	bm     health.BeltManager
}

func NewBuilder(logger *zap.Logger, sm town.ShopManager, bm health.BeltManager) Builder {
	return Builder{
		logger: logger,
		sm:     sm,
		bm:     bm,
	}
}
