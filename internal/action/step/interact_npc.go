package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

type InteractNPC struct {
	basicStep
	NPC                   game.NPCID
	waitingForInteraction bool
}

func NewInteractNPC(npc game.NPCID) *InteractNPC {
	return &InteractNPC{
		basicStep: basicStep{
			status: StatusNotStarted,
		},
		NPC: npc,
	}
}

func (i *InteractNPC) Status(data game.Data) Status {
	// Give some extra time to render the UI
	if data.OpenMenus.NPCInteract && time.Since(i.lastRun) > time.Second*1 {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractNPC) Run(data game.Data) error {
	i.tryTransitionStatus(StatusInProgress)
	if time.Since(i.lastRun) < time.Millisecond*500 {
		return nil
	}

	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second*3 {
		return nil
	}

	i.lastRun = time.Now()
	m, found := data.Monsters[i.NPC]
	if found {
		if m.IsHovered {
			hid.Click(hid.LeftButton)
			i.waitingForInteraction = true
			return nil
		}
	}

	x, y := i.getNPCPosition(data)

	// TODO: Handle not found
	path, distance, _ := helper.GetPathToDestination(data, x, y)
	if distance > 15 {
		helper.MoveThroughPath(path, 15, false)
		return nil
	}
	helper.MoveThroughPath(path, 0, false)

	return nil
}

func (i InteractNPC) getNPCPosition(d game.Data) (X, Y int) {
	npc, found := d.Monsters[i.NPC]
	if found {
		// Position is bottom hitbox by default, let's move it a bit
		return npc.Position.X - 2, npc.Position.Y - 2
	}

	return d.NPCs[i.NPC].Positions[0].X, d.NPCs[i.NPC].Positions[0].Y
}
