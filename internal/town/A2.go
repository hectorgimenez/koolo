package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type A2 struct {
}

func (a A2) MercContractorNPC() game.NPCID {
	return game.GreizNPC
}

func (a A2) RefillNPC() game.NPCID {
	return game.DrognanNPC
}

func (a A2) RepairNPC() game.NPCID {
	return game.FaraNPC
}
