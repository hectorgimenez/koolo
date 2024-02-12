package action

import (
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/town"
	"log/slog"
)

type Builder struct {
	logger *slog.Logger
	sm     town.ShopManager
	bm     health.BeltManager
	gr     *reader.GameReader
	ch     Character
}

func NewBuilder(logger *slog.Logger, sm town.ShopManager, bm health.BeltManager, gr *reader.GameReader, ch Character) *Builder {
	return &Builder{
		logger: logger,
		sm:     sm,
		bm:     bm,
		gr:     gr,
		ch:     ch,
	}
}
