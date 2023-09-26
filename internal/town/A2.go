package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
)

type A2 struct {
}

func (a A2) GamblingNPC() npc.ID {
	return npc.Elzix
}

func (a A2) HealNPC() npc.ID {
	return npc.Fara
}

func (a A2) MercContractorNPC() npc.ID {
	return npc.Greiz
}

func (a A2) RefillNPC() npc.ID {
	return npc.Drognan
}

func (a A2) RepairNPC() npc.ID {
	return npc.Fara
}

func (a A2) TPWaitingArea(_ data.Data) data.Position {
	return data.Position{
		X: 5161,
		Y: 5059,
	}
}
