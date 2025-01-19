package step

import (
	"errors"
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/ui"
)

func InteractNPC(npcID npc.ID) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractNPC")

	const (
		maxAttempts        = 8
		interactionTimeout = 3 * time.Second
		minMenuOpenWait    = 300 * time.Millisecond
		maxDistance        = 15
		hoverTimeout       = 800 * time.Millisecond
	)

	var lastInteractionTime time.Time
	var targetNPCID data.UnitID

	for attempts := 0; attempts < maxAttempts; attempts++ {
		ctx.PauseIfNotPriority()

		// Clear last interaction if we've waited too long
		if !lastInteractionTime.IsZero() && time.Since(lastInteractionTime) > interactionTimeout {
			lastInteractionTime = time.Time{}
			targetNPCID = 0
		}

		// Check if interaction succeeded and menu is open
		if ctx.Data.OpenMenus.NPCInteract || ctx.Data.OpenMenus.NPCShop {
			// Find current NPC position
			if targetNPCID != 0 {
				if currentNPC, found := ctx.Data.Monsters.FindByID(targetNPCID); found {
					currentDistance := pather.DistanceFromPoint(currentNPC.Position, ctx.Data.PlayerUnit.Position)
					if currentDistance <= maxDistance {
						// Success - wait minimum time for menu to fully open
						time.Sleep(minMenuOpenWait)
						return nil
					}
				}
			}

			// Wrong NPC, too far, or NPC moved away - close menu and retry
			CloseAllMenus()
			time.Sleep(200 * time.Millisecond)
			targetNPCID = 0
			continue
		}

		// Don't attempt new interaction if we're waiting for previous one
		if !lastInteractionTime.IsZero() && time.Since(lastInteractionTime) < minMenuOpenWait {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Find and validate target NPC
		townNPC, found := ctx.Data.Monsters.FindOne(npcID, data.MonsterTypeNone)
		if !found {
			if attempts == maxAttempts-1 {
				return fmt.Errorf("NPC %d not found after %d attempts", npcID, maxAttempts)
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		distance := ctx.PathFinder.DistanceFromMe(townNPC.Position)
		if distance > maxDistance {
			return fmt.Errorf("NPC %d is too far away (distance: %d)", npcID, distance)
		}

		// Calculate screen coordinates based on current NPC position
		x, y := ui.GameCoordsToScreenCords(townNPC.Position.X, townNPC.Position.Y)
		// Special case for Tyrael's hitbox
		if npcID == npc.ID(240) {
			y = y - 40
		}

		ctx.HID.MovePointer(x, y)

		// Wait for hover before clicking
		hoverWaitStart := time.Now()
		hoverFound := false
		var hoveredNPC data.Monster

		for time.Since(hoverWaitStart) < hoverTimeout {
			// Get fresh NPC position in case they moved
			if currentNPC, found := ctx.Data.Monsters.FindOne(npcID, data.MonsterTypeNone); found {
				if currentNPC.IsHovered {
					hoveredNPC = currentNPC
					hoverFound = true
					break
				}

				// Update mouse position if NPC moved
				newX, newY := ui.GameCoordsToScreenCords(currentNPC.Position.X, currentNPC.Position.Y)
				if newX != x || newY != y {
					if npcID == npc.ID(240) {
						newY = newY - 40
					}
					ctx.HID.MovePointer(newX, newY)
					x, y = newX, newY
				}
			}
			time.Sleep(50 * time.Millisecond)
		}

		if !hoverFound {
			continue
		}

		// Store the NPC ID we're interacting with
		targetNPCID = hoveredNPC.UnitID
		ctx.HID.Click(game.LeftButton, x, y)
		lastInteractionTime = time.Now()

		// Wait a bit for the menu to open
		time.Sleep(minMenuOpenWait)
	}

	return errors.New("failed to interact with NPC after all attempts")
}
