package town

import (
	"github.com/hectorgimenez/koolo/internal/game/data"
)

type A5 struct {
}

func (a A5) MercContractorNPC() data.NPCID {
	return data.QualKehkNPC
}

func (a A5) RefillNPC() data.NPCID {
	return data.MalahNPC
}

func (a A5) RepairNPC() data.NPCID {
	return data.LarzukNPC
}
