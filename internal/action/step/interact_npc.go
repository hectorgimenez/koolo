package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func InteractNPC(npcID npc.ID) error {
	maxInteractionAttempts := 5
	interactionAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	ctx := context.Get()
	ctx.ContextDebug.LastStep = "InteractNPC"

	for {
		ctx.RefreshGameData()

		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		if ctx.Data.OpenMenus.NPCInteract {
			return nil
		}

		if interactionAttempts >= maxInteractionAttempts {
			return errors.New("failed interacting with NPC")
		}

		// Give some time before retrying the interaction
		if waitingForInteraction && time.Since(lastRun) < time.Millisecond*500 {
			continue
		}

		lastRun = time.Now()
		m, found := ctx.Data.Monsters.FindOne(npcID, data.MonsterTypeNone)
		if found {
			if m.IsHovered {
				ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
				waitingForInteraction = true
				interactionAttempts++
				continue
			}

			distance := ctx.PathFinder.DistanceFromMe(m.Position)
			if distance > 15 {
				return fmt.Errorf("NPC is too far away: %d. Current distance: %d", npcID, distance)
			}

			x, y := ui.GameCoordsToScreenCords(m.Position.X, m.Position.Y)
			// Act 4 Tyrael has a super weird hitbox
			if npcID == npc.Tyrael2 {
				y = y - 40
			}
			currentMouseCoords = data.Position{X: x, Y: y}
			ctx.HID.MovePointer(x, y)
			interactionAttempts++
		}
	}
}
