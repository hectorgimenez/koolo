package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
)

type A5 struct {
}

func (a A5) GamblingNPC() npc.ID {
	return npc.Drehya // Anya
}

func (a A5) HealNPC() npc.ID {
	return npc.Malah
}

func (a A5) MercContractorNPC() npc.ID {
	return npc.QualKehk
}

func (a A5) RefillNPC() npc.ID {
	return npc.Malah
}

func (a A5) RepairNPC() npc.ID {
	return npc.Larzuk
}

func (a A5) TPWaitingArea(_ data.Data) data.Position {
	return data.Position{
		X: 5104,
		Y: 5019,
	}
}
