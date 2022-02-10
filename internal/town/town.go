package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type Town interface {
	RefillNPC() game.NPCID
	RepairNPC() game.NPCID
	MercContractorNPC() game.NPCID
}
