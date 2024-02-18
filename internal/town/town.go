package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
)

func getTowns() map[area.Area]Town {
	return map[area.Area]Town{
		area.RogueEncampment:        A1{},
		area.LutGholein:             A2{},
		area.KurastDocks:            A3{},
		area.ThePandemoniumFortress: A4{},
		area.Harrogath:              A5{},
	}
}

type Town interface {
	RefillNPC() npc.ID
	HealNPC() npc.ID
	RepairNPC() npc.ID
	MercContractorNPC() npc.ID
	GamblingNPC() npc.ID
	TPWaitingArea(d data.Data) data.Position
}

func GetTownByArea(a area.Area) Town {
	return getTowns()[a]
}

func IsTown(a area.Area) bool {
	for aa, _ := range getTowns() {
		if aa == a {
			return true
		}
	}
	return false
}
