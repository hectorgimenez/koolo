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
	ctx := context.Get()
	ctx.SetLastStep("InteractNPC")

	const (
		maxAttempts        = 5
		interactionTimeout = 3 * time.Second
		minMenuOpenWait    = 200 * time.Millisecond
	)

	var lastInteractionTime time.Time
	var currentMouseCoords data.Position

	for attempts := 0; attempts < maxAttempts; attempts++ {
		ctx.RefreshGameData()
		ctx.PauseIfNotPriority()

		// Clear last interaction if we've waited long enough
		if !lastInteractionTime.IsZero() && time.Since(lastInteractionTime) > interactionTimeout {
			lastInteractionTime = time.Time{}
		}

		// Check if interaction succeeded
		if ctx.Data.OpenMenus.NPCInteract {
			// Verify we're interacting with the right NPC by checking distance
			if townNPC, found := ctx.Data.Monsters.FindOne(npcID, data.MonsterTypeNone); found {
				if ctx.PathFinder.DistanceFromMe(townNPC.Position) <= 15 {
					// Wait a minimum time to ensure menu is fully open
					time.Sleep(minMenuOpenWait)
					return nil
				}
			}
			// Wrong NPC or too far - close menu and retry
			ctx.HID.PressKey(0x1B) // ESC
			time.Sleep(200 * time.Millisecond)
			continue
		}

		// Don't attempt new interaction if we're waiting for previous one
		if !lastInteractionTime.IsZero() && time.Since(lastInteractionTime) < minMenuOpenWait {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Find and interact with NPC
		townNPC, found := ctx.Data.Monsters.FindOne(npcID, data.MonsterTypeNone)
		if !found {
			if attempts == maxAttempts-1 {
				return fmt.Errorf("NPC %d not found after %d attempts", npcID, maxAttempts)
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		distance := ctx.PathFinder.DistanceFromMe(townNPC.Position)
		if distance > 15 {
			return fmt.Errorf("NPC %d is too far away (distance: %d)", npcID, distance)
		}

		x, y := ui.GameCoordsToScreenCords(townNPC.Position.X, townNPC.Position.Y)
		// Act 4 Tyrael has a super weird hitbox
		if npcID == npc.ID(240) {
			y = y - 40
		}

		currentMouseCoords = data.Position{X: x, Y: y}
		ctx.HID.MovePointer(x, y)

		// Wait for hover
		hoverWaitStart := time.Now()
		hoverFound := false
		for time.Since(hoverWaitStart) < 500*time.Millisecond {
			ctx.RefreshGameData()
			if townNPC, found := ctx.Data.Monsters.FindOne(npcID, data.MonsterTypeNone); found && townNPC.IsHovered {
				hoverFound = true
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		if !hoverFound {
			continue
		}

		ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
		lastInteractionTime = time.Now()
	}

	return errors.New("failed to interact with NPC after all attempts")
}
