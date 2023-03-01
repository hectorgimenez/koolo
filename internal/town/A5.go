package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
)

type A5 struct {
}

func (a A5) HealNPC() npc.ID {
	return npc.Malah
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

func (a A5) TPWaitingArea(_ game.Data) game.Position {
	return game.Position{
		X: 5104,
		Y: 5019,
	}
}
