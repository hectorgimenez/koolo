package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type InteractNPCStep struct {
	pathingStep
	NPC                   npc.ID
	waitingForInteraction bool
	isCompletedFn         func(d data.Data) bool
}

func InteractNPC(npc npc.ID) *InteractNPCStep {
	return &InteractNPCStep{
		pathingStep: newPathingStep(),
		NPC:         npc,
	}
}

func InteractNPCWithCheck(npc npc.ID, isCompletedFn func(d data.Data) bool) *InteractNPCStep {
	return &InteractNPCStep{
		pathingStep:   newPathingStep(),
		NPC:           npc,
		isCompletedFn: isCompletedFn,
	}
}

func (i *InteractNPCStep) Status(d data.Data) Status {
	if i.status == StatusCompleted {
		return StatusCompleted
	}

	if i.isCompletedFn != nil && i.isCompletedFn(d) {
		return i.tryTransitionStatus(StatusCompleted)
	}

	// Give some extra time to render the UI
	if d.OpenMenus.NPCInteract && time.Since(i.lastRun) > time.Second*1 {
		return i.tryTransitionStatus(StatusCompleted)
	}

	return i.status
}

func (i *InteractNPCStep) Run(d data.Data) error {
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
	m, found := d.Monsters.FindOne(i.NPC, data.MonsterTypeNone)
	if found {
		if m.IsHovered {
			hid.Click(hid.LeftButton)
			i.waitingForInteraction = true
			return nil
		}
	}

	pos, found := i.getNPCPosition(d)
	if !found {
		return fmt.Errorf("NPC not found")
	}

	distance := pather.DistanceFromMe(d, pos)
	if distance > 15 {
		path, _, found := pather.GetPath(d, pos)
		if !found {
			pather.RandomMovement()
			i.consecutivePathNotFound++
			return nil
		}
		i.consecutivePathNotFound = 0
		pather.MoveThroughPath(path, helper.RandRng(7, 17), false)
		return nil
	}
	x, y := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, pos.X, pos.Y)
	hid.MovePointer(x, y)

	return nil
}

func (i *InteractNPCStep) getNPCPosition(d data.Data) (data.Position, bool) {
	monster, found := d.Monsters.FindOne(i.NPC, data.MonsterTypeNone)
	if found {
		// Position is bottom hitbox by default, let's move it a bit
		return data.Position{X: monster.Position.X - 2, Y: monster.Position.Y - 2}, true
	}

	npc, found := d.NPCs.FindOne(i.NPC)
	if !found {
		return data.Position{}, false
	}

	return npc.Positions[0], true
}
