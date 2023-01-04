package action

import (
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/memory"
	"github.com/hectorgimenez/koolo/internal/town"
	"go.uber.org/zap"
)

type Builder struct {
	logger *zap.Logger
	sm     town.ShopManager
	bm     health.BeltManager
	gr     *memory.GameReader
	ch     Character
}

func NewBuilder(logger *zap.Logger, sm town.ShopManager, bm health.BeltManager, gr *memory.GameReader, ch Character) Builder {
	return Builder{
		logger: logger,
		sm:     sm,
		bm:     bm,
		gr:     gr,
		ch:     ch,
	}
}
