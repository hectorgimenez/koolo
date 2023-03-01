package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
)

type A4 struct {
}

func (a A4) HealNPC() npc.ID {
	return npc.Jamella
}

func (a A4) MercContractorNPC() npc.ID {
	return npc.Tyrael2
}

func (a A4) RefillNPC() npc.ID {
	return npc.Jamella
}

func (a A4) RepairNPC() npc.ID {
	return npc.Halbu
}

func (a A4) TPWaitingArea(_ game.Data) game.Position {
	return game.Position{
		X: 5047,
		Y: 5033,
	}
}
