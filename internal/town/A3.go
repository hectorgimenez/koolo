package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type A3 struct {
}

func (a A3) MercContractorNPC() game.NPCID {
	return game.AshearaNPC
}

func (a A3) RefillNPC() game.NPCID {
	return game.OrmusNPC
}

func (a A3) RepairNPC() game.NPCID {
	return game.HratliNPC
}
