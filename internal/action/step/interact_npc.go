package step

import (
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
		maxAttempts     = 8
		minMenuOpenWait = 300 * time.Millisecond
		maxDistance     = 15
		hoverWait       = 800 * time.Millisecond
	)

	var targetNPCID data.UnitID

	for attempts := 0; attempts < maxAttempts; attempts++ {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		// Check if interaction succeeded and menu is open
		if ctx.Data.OpenMenus.NPCInteract || ctx.Data.OpenMenus.NPCShop {
			// Find current NPC position
			if targetNPCID != 0 {
				if currentNPC, found := ctx.Data.Monsters.FindByID(targetNPCID); found {
					currentDistance := pather.DistanceFromPoint(currentNPC.Position, ctx.Data.PlayerUnit.Position)
					if currentDistance <= maxDistance {
						time.Sleep(minMenuOpenWait)
						return nil
					}
				}
			}

			// Wrong NPC, too far, or NPC moved - close menu and retry
			CloseAllMenus()
			time.Sleep(200 * time.Millisecond)
			targetNPCID = 0
			continue
		}

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

		// Calculate click position
		x, y := ui.GameCoordsToScreenCords(townNPC.Position.X, townNPC.Position.Y)
		if npcID == npc.Tyrael2 || npcID == npc.Tyrael {
			y = y - 40 // Tyrael has a super weird hitbox
		}

		// Move mouse and wait for hover
		ctx.HID.MovePointer(x, y)
		hoverStart := time.Now()

		for time.Since(hoverStart) < hoverWait {
			if currentNPC, found := ctx.Data.Monsters.FindOne(npcID, data.MonsterTypeNone); found && currentNPC.IsHovered {
				targetNPCID = currentNPC.UnitID
				ctx.HID.Click(game.LeftButton, x, y)
				time.Sleep(minMenuOpenWait)
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	return fmt.Errorf("failed to interact with NPC after %d attempts", maxAttempts)
}
