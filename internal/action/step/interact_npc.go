package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type InteractNPCStep struct {
	pathingStep
	NPC                   game.NPCID
	waitingForInteraction bool
}

func InteractNPC(npc game.NPCID) *InteractNPCStep {
	return &InteractNPCStep{
		pathingStep: newPathingStep(),
		NPC:         npc,
	}
}

func (i *InteractNPCStep) Status(data game.Data) Status {
	// Give some extra time to render the UI
	if data.OpenMenus.NPCInteract && time.Since(i.lastRun) > time.Second*1 {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractNPCStep) Run(data game.Data) error {
	if i.consecutivePathNotFound >= maxPathNotFoundRetries {
		return fmt.Errorf("error moving to %s: %w", i.NPC, errPathNotFound)
	}

	i.tryTransitionStatus(StatusInProgress)
	// Throttle movement clicks
	if time.Since(i.lastRun) < time.Millisecond*350 {
		return nil
	}

	// Give some time before retrying the interaction
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second*2 {
		return nil
	}

	i.lastRun = time.Now()
	m, found := data.Monsters.FindOne(i.NPC)
	if found {
		if m.IsHovered {
			hid.Click(hid.LeftButton)
			i.waitingForInteraction = true
			return nil
		}
	}

	x, y, found := i.getNPCPosition(data)
	if !found {
		return fmt.Errorf("NPC not found")
	}

	distance := pather.DistanceFromPoint(data, x, y)
	if distance > 15 {
		path, _, found := pather.GetPathToDestination(data, x, y)
		if !found {
			pather.RandomMovement()
			i.consecutivePathNotFound++
			return nil
		}
		i.consecutivePathNotFound = 0
		pather.MoveThroughPath(path, 12, false)
		return nil
	}
	x, y = pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, x, y)
	hid.MovePointer(x, y)

	return nil
}

func (i InteractNPCStep) getNPCPosition(d game.Data) (X, Y int, found bool) {
	monster, found := d.Monsters.FindOne(i.NPC)
	if found {
		// Position is bottom hitbox by default, let's move it a bit
		return monster.Position.X - 2, monster.Position.Y - 2, true
	}

	npc, found := d.NPCs.FindOne(i.NPC)
	if !found {
		return 0, 0, false
	}

	return npc.Positions[0].X, npc.Positions[0].Y, true
}
