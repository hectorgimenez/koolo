package item

import (
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"go.uber.org/zap"
)

type Finder struct {
	logger    *zap.Logger
	dr        data.DataRepository
	pickitCfg config.Pickit
}

func (f Finder) getItemsToPickup() {

}
