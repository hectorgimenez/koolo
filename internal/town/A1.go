package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type A1 struct {
}

func (a A1) MercContractorNPC() game.NPCID {
	return game.KashyaNPC
}

func (a A1) RefillNPC() game.NPCID {
	return game.AkaraNPC
}

func (a A1) RepairNPC() game.NPCID {
	return game.CharsiNPC
}
