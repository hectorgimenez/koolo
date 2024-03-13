package action

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/town"
)

type Builder struct {
	sm town.ShopManager
	bm health.BeltManager
	ch Character
	container.Container
}

func NewBuilder(container container.Container, sm town.ShopManager, bm health.BeltManager, ch Character) *Builder {
	return &Builder{
		sm:        sm,
		bm:        bm,
		ch:        ch,
		Container: container,
	}
}
