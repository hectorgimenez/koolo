package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type InteractNPCStep struct {
	basicStep
	NPC                   npc.ID
	waitingForInteraction bool
	isCompletedFn         func(d data.Data) bool
}

func InteractNPC(npc npc.ID) *InteractNPCStep {
	return &InteractNPCStep{
		basicStep: newBasicStep(),
		NPC:       npc,
	}
}

func InteractNPCWithCheck(npc npc.ID, isCompletedFn func(d data.Data) bool) *InteractNPCStep {
	return &InteractNPCStep{
		basicStep:     newBasicStep(),
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
	i.tryTransitionStatus(StatusInProgress)

	// Give some time before retrying the interaction
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second {
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

		distance := pather.DistanceFromMe(d, m.Position)
		if distance > 15 {
			return fmt.Errorf("NPC is too far away: %d. Current distance: %d", i.NPC, distance)
		}

		x, y := pather.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, m.Position.X, m.Position.Y)

		// Act 4 Tyrael has a super weird hitbox
		if i.NPC == npc.Tyrael2 {
			hid.MovePointer(x, y-20)
		} else {
			hid.MovePointer(x, y)
		}

		return nil
	}

	return fmt.Errorf("npc %d not found", i.NPC)
}
