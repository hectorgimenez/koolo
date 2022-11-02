package town

import (
	"github.com/hectorgimenez/koolo/internal/game/npc"
)

type A1 struct {
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
