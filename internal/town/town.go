package town

import "github.com/hectorgimenez/koolo/internal/game/data"

type Town interface {
	RefillNPC() data.NPCID
	RepairNPC() data.NPCID
	MercContractorNPC() data.NPCID
}
