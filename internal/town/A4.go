package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/game"
)

type A4 struct {
}

func (a A4) GamblingNPC() npc.ID {
	return npc.Jamella
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

func (a A4) IdentifyNPC() npc.ID {
	return npc.DeckardCain4
}

func (a A4) TPWaitingArea(_ game.Data) data.Position {
	return data.Position{
		X: 5047,
		Y: 5033,
	}
}

func (a A4) TownArea() area.ID {
	return area.ThePandemoniumFortress
}
