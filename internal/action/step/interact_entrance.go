package step

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
	"time"
)

const (
	maxEntranceDistance = 6
)

func InteractEntrance(area area.ID) error {
	maxInteractionAttempts := 5
	interactionAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		// Give some extra time to render the UI
		if ctx.Data.AreaData.Area == area && time.Since(lastRun) > time.Millisecond*500 && ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) {
			// We've successfully entered the new area
			return nil
		}

		if interactionAttempts > maxInteractionAttempts {
			return fmt.Errorf("area %s [%d] could not be interacted", area.Area().Name, area)
		}

		if waitingForInteraction && time.Since(lastRun) < time.Millisecond*500 {
			continue
		}

		lastRun = time.Now()
		for _, l := range ctx.Data.AdjacentLevels {
			if l.Area == area {
				distance := ctx.PathFinder.DistanceFromMe(l.Position)
				if distance > maxEntranceDistance {
					return fmt.Errorf("entrance too far away (distance: %d)", distance)
				}

				if l.IsEntrance {
					// Adjust click position to be slightly closer to entrance
					lx, ly := ctx.PathFinder.GameCoordsToScreenCords(l.Position.X-1, l.Position.Y-1)
					if ctx.Data.HoverData.UnitType == 5 || ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered {
						ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
						waitingForInteraction = true
						utils.Sleep(200) // Small delay after click
					}

					x, y := utils.Spiral(interactionAttempts)
					x = x / 3 // Reduce spiral size further
					y = y / 3
					currentMouseCoords = data.Position{X: lx + x, Y: ly + y}
					ctx.HID.MovePointer(lx+x, ly+y)
					interactionAttempts++
					utils.Sleep(100) // Small delay for mouse movement
					continue
				}

				return fmt.Errorf("area %s [%d] is not an entrance", area.Area().Name, area)
			}
		}
	}
}
