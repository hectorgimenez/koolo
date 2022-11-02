package town

import (
	"github.com/hectorgimenez/koolo/internal/game/npc"
)

type A3 struct {
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
