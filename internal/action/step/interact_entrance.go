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

const maxEntranceDistance = 10

func InteractEntrance(dest area.ID) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	interactionAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastInteraction := time.Time{}

	for {
		ctx.PauseIfNotPriority()

		// Success check - we've entered the new area
		if ctx.Data.PlayerUnit.Area == dest {
			return nil
		}

		if interactionAttempts >= maxInteractionAttempts {
			return fmt.Errorf("failed to enter area %s after %d attempts", dest.Area().Name, maxInteractionAttempts)
		}

		// If waiting for interaction, give some time before retrying
		if waitingForInteraction && time.Since(lastInteraction) < time.Millisecond*500 {
			utils.Sleep(50)
			continue
		}

		// Reset waiting state if enough time has passed
		if waitingForInteraction && time.Since(lastInteraction) >= time.Millisecond*500 {
			waitingForInteraction = false
		}

		// Find the entrance we want to interact with
		for _, l := range ctx.Data.AdjacentLevels {
			if l.Area == dest {
				// Verify distance
				if ctx.PathFinder.DistanceFromMe(l.Position) > maxEntranceDistance {
					return fmt.Errorf("entrance is too far away")
				}

				if l.IsEntrance {
					lx, ly := ctx.PathFinder.GameCoordsToScreenCords(l.Position.X-2, l.Position.Y-2)

					// Check if we're hovering over the entrance
					if ctx.Data.HoverData.UnitType == 5 || (ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered) {
						ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
						waitingForInteraction = true
						lastInteraction = time.Now()
						utils.Sleep(200)
						continue
					}

					// Only try new mouse positions if not waiting for interaction
					if !waitingForInteraction {
						x, y := utils.Spiral(interactionAttempts)
						currentMouseCoords = data.Position{X: lx + x, Y: ly + y}
						ctx.HID.MovePointer(lx+x, ly+y)
						interactionAttempts++
						utils.Sleep(100)
					}
					continue
				}

				return fmt.Errorf("area %s [%d] is not an entrance", l.Area.Area().Name, l.Area)
			}
		}

		// If we get here without finding the entrance, wait briefly and refresh
		utils.Sleep(100)
		ctx.RefreshGameData()
	}
}
