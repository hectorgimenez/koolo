package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type InteractNPCStep struct {
	basicStep
	NPC                   npc.ID
	waitingForInteraction bool
	isCompletedFn         func(d game.Data) bool
	currentMouseCoords    data.Position
}

func InteractNPC(npc npc.ID) *InteractNPCStep {
	return &InteractNPCStep{
		basicStep: newBasicStep(),
		NPC:       npc,
	}
}

func InteractNPCWithCheck(npc npc.ID, isCompletedFn func(d game.Data) bool) *InteractNPCStep {
	return &InteractNPCStep{
		basicStep:     newBasicStep(),
		NPC:           npc,
		isCompletedFn: isCompletedFn,
	}
}

func (i *InteractNPCStep) Status(d game.Data, _ container.Container) Status {
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

func (i *InteractNPCStep) Run(d game.Data, container container.Container) error {
	i.tryTransitionStatus(StatusInProgress)

	// Give some time before retrying the interaction
	if i.waitingForInteraction && time.Since(i.lastRun) < time.Second {
		return nil
	}

	i.lastRun = time.Now()
	m, found := d.Monsters.FindOne(i.NPC, data.MonsterTypeNone)
	if found {
		if m.IsHovered {
			container.HID.Click(game.LeftButton, i.currentMouseCoords.X, i.currentMouseCoords.Y)
			i.waitingForInteraction = true
			return nil
		}

		distance := pather.DistanceFromMe(d, m.Position)
		if distance > 15 {
			return fmt.Errorf("NPC is too far away: %d. Current distance: %d", i.NPC, distance)
		}

		x, y := container.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, m.Position.X, m.Position.Y)
		// Act 4 Tyrael has a super weird hitbox
		if i.NPC == npc.Tyrael2 {
			y = y - 40
		}
		i.currentMouseCoords = data.Position{X: x, Y: y}
		container.HID.MovePointer(x, y)

		return nil
	}

	return fmt.Errorf("npc %d not found", i.NPC)
}
