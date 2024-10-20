package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/game"
)

type A3 struct {
}

func (a A3) GamblingNPC() npc.ID {
	return npc.Alkor
}

func (a A3) HealNPC() npc.ID {
	return npc.Ormus
}

func (a A3) MercContractorNPC() npc.ID {
	return npc.Asheara
}

func (a A3) RefillNPC() npc.ID {
	return npc.Ormus
}

func (a A3) RepairNPC() npc.ID {
	return npc.Hratli
}

func (a A3) IdentifyNPC() npc.ID {
	return npc.DeckardCain3
}

func (a A3) TPWaitingArea(_ game.Data) data.Position {
	return data.Position{
		X: 5151,
		Y: 5068,
	}
}

func (a A3) TownArea() area.ID {
	return area.KurastDocks
}
