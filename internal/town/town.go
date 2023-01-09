package town

import (
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/npc"
)

type Town interface {
	RefillNPC() npc.ID
	HealNPC() npc.ID
	RepairNPC() npc.ID
	MercContractorNPC() npc.ID
}

func GetTownByArea(a area.Area) Town {
	towns := map[area.Area]Town{
		area.RogueEncampment:        A1{},
		area.LutGholein:             A2{},
		area.KurastDocks:            A3{},
		area.ThePandemoniumFortress: A4{},
		area.Harrogath:              A5{},
	}

	return towns[a]
}
