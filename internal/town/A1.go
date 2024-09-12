package town

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/game"
)

type A1 struct {
}

func (a A1) GamblingNPC() npc.ID {
	return npc.Gheed
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

func (a A1) IdentifyNPC() npc.ID {
	return npc.DeckardCain5
}

func (a A1) TPWaitingArea(d game.Data) data.Position {
	rogueBonfire, found := d.Objects.FindOne(object.RogueBonfire)
	if found {
		return data.Position{
			X: rogueBonfire.Position.X + 10,
			Y: rogueBonfire.Position.Y + 10,
		}
	}

	cain, _ := d.NPCs.FindOne(npc.Kashya)

	return cain.Positions[0]
}

func (a A1) TownArea() area.ID {
	return area.RogueEncampment
}
