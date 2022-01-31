package town

import (
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type A5 struct {
	dr data.DataRepository
	pf helper.PathFinder
}

func (a A5) OpenWP() {
}

func (a A5) OpenVendorTrade() {
	a.pf.InteractToNPC(data.MalahNPC)
}
