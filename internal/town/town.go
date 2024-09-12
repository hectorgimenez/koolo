package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Town interface {
	RefillNPC() npc.ID
	HealNPC() npc.ID
	RepairNPC() npc.ID
	MercContractorNPC() npc.ID
	GamblingNPC() npc.ID
	IdentifyNPC() npc.ID
	TPWaitingArea(d game.Data) data.Position
	TownArea() area.ID
}

func GetTownByArea(a area.ID) Town {
	switch a.Act() {
	case 1:
		return A1{}
	case 2:
		return A2{}
	case 3:
		return A3{}
	case 4:
		return A4{}
	}

	return A5{}
}
