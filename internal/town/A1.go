package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
)

type A1 struct {
}

func (a A1) GamblingNPC() npc.ID {
	return npc.Gheed
}

func (a A1) HealNPC() npc.ID {
	return npc.Akara
}

func (a A1) MercContractorNPC() npc.ID {
	return npc.Kashya
}

func (a A1) RefillNPC() npc.ID {
	return npc.Akara
}

func (a A1) RepairNPC() npc.ID {
	return npc.Charsi
}

func (a A1) TPWaitingArea(d data.Data) data.Position {
	cain, _ := d.NPCs.FindOne(npc.Kashya)

	return cain.Positions[0]
}
