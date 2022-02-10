package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type A5 struct {
}

func (a A5) MercContractorNPC() game.NPCID {
	return game.QualKehkNPC
}

func (a A5) RefillNPC() game.NPCID {
	return game.MalahNPC
}

func (a A5) RepairNPC() game.NPCID {
	return game.LarzukNPC
}
