package action

import (
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/reader"
	"github.com/hectorgimenez/koolo/internal/town"
	"github.com/hectorgimenez/koolo/internal/ui"
	"go.uber.org/zap"
)

type Builder struct {
	logger *zap.Logger
	sm     town.ShopManager
	bm     health.BeltManager
	gr     *reader.GameReader
	ch     Character
	tf     *ui.TemplateFinder
}

func NewBuilder(logger *zap.Logger, sm town.ShopManager, bm health.BeltManager, gr *reader.GameReader, ch Character, tf *ui.TemplateFinder) *Builder {
	return &Builder{
		logger: logger,
		sm:     sm,
		bm:     bm,
		gr:     gr,
		ch:     ch,
		tf:     tf,
	}
}
