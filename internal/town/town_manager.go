package town

import (
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Manager struct {
	towns map[data.Area]Town
	pf    helper.PathFinder
}

func NewTownManager(repository data.DataRepository, pf helper.PathFinder) Manager {
	return Manager{
		towns: map[data.Area]Town{
			data.AreaHarrogath: A5{
				dr: repository,
				pf: pf,
			},
		},
	}
}

func (tm Manager) BuyPotionsAndTPs(area data.Area) {
	t := tm.getTownByArea(area)
	t.OpenVendorTrade()
}

func (tm Manager) getTownByArea(area data.Area) Town {
	return tm.towns[area]
}
