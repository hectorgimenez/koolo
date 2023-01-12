package step

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"time"
)

type InteractNPCStep struct {
	pathingStep
	NPC                   npc.ID
	waitingForInteraction bool
}

func InteractNPC(npc npc.ID) *InteractNPCStep {
	return &InteractNPCStep{
		pathingStep: newPathingStep(),
		NPC:         npc,
	}
}

func (i *InteractNPCStep) Status(data game.Data) Status {
	if i.status == StatusCompleted {
		return StatusCompleted
	}

	// Give some extra time to render the UI
	if data.OpenMenus.NPCInteract && time.Since(i.lastRun) > time.Second*1 {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractNPCStep) Run(data game.Data) error {
	// Throttle movement clicks
	if time.Since(i.lastRun) < helper.RandomDurationMs(300, 600) {
		return nil
	}

	if i.consecutivePathNotFound >= maxPathNotFoundRetries {
		return fmt.Errorf("error moving to %s: %w", i.NPC, errPathNotFound)
	}

	i.tryTransitionStatus(StatusInProgress)

	// Give some time before retrying the interaction
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second*2 {
		return nil
	}

	i.lastRun = time.Now()
	m, found := data.Monsters.FindOne(i.NPC, game.MonsterTypeNone)
	if found {
		if m.IsHovered {
			hid.Click(hid.LeftButton)
			i.waitingForInteraction = true
			return nil
		}
	}

	pos, found := i.getNPCPosition(data)
	if !found {
		return fmt.Errorf("NPC not found")
	}

	distance := pather.DistanceFromMe(data, pos)
	if distance > 15 {
		path, _, found := pather.GetPath(data, pos.X, pos.Y)
		if !found {
			pather.RandomMovement()
			i.consecutivePathNotFound++
			return nil
		}
		i.consecutivePathNotFound = 0
		pather.MoveThroughPath(path, helper.RandRng(7, 17), false)
		return nil
	}
	x, y := pather.GameCoordsToScreenCords(data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y, pos.X, pos.Y)
	hid.MovePointer(x, y)

	return nil
}

func (i *InteractNPCStep) getNPCPosition(d game.Data) (game.Position, bool) {
	monster, found := d.Monsters.FindOne(i.NPC, game.MonsterTypeNone)
	if found {
		// Position is bottom hitbox by default, let's move it a bit
		return game.Position{X: monster.Position.X - 2, Y: monster.Position.Y - 2}, true
	}

	npc, found := d.NPCs.FindOne(i.NPC)
	if !found {
		return game.Position{}, false
	}

	return npc.Positions[0], true
}
