package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
)

type Town interface {
	RefillNPC() npc.ID
	HealNPC() npc.ID
	RepairNPC() npc.ID
	MercContractorNPC() npc.ID
	GamblingNPC() npc.ID
	TPWaitingArea(d data.Data) data.Position
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
