package town

import (
	"github.com/hectorgimenez/koolo/internal/game/npc"
)

type A5 struct {
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
