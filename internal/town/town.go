package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type Town interface {
	RefillNPC() game.NPCID
	RepairNPC() game.NPCID
	MercContractorNPC() game.NPCID
}

func GetTownByArea(area game.Area) Town {
	towns := map[game.Area]Town{
		game.AreaRogueEncampment: A1{},
		game.AreaKurastDocks:     A3{},
		game.AreaHarrogath:       A5{},
	}

	return towns[area]
}
