package town

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
)

type A1 struct {
}

func (a A1) HealNPC() npc.ID {
	return npc.Akara
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

func (a A1) TPWaitingArea(d game.Data) game.Position {
	cain, _ := d.NPCs.FindOne(npc.Kashya)

	return cain.Positions[0]
}
