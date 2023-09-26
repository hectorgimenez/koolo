package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
)

type A4 struct {
}

func (a A4) GamblingNPC() npc.ID {
	return npc.Jamella
}

func (a A4) HealNPC() npc.ID {
	return npc.Jamella
}

func (a A4) MercContractorNPC() npc.ID {
	return npc.Tyrael2
}

func (a A4) RefillNPC() npc.ID {
	return npc.Jamella
}

func (a A4) RepairNPC() npc.ID {
	return npc.Halbu
}

func (a A4) TPWaitingArea(_ data.Data) data.Position {
	return data.Position{
		X: 5047,
		Y: 5033,
	}
}
