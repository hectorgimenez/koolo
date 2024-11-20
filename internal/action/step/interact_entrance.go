package step

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/ui"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const maxEntranceDistance = 6

func InteractEntrance(dest area.ID) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	interactionAttempts := 0
	currentMouseCoords := data.Position{}

	for {
		ctx.PauseIfNotPriority()

		// Success check - we've entered the new area
		if ctx.Data.PlayerUnit.Area == dest {
			return nil
		}

		if interactionAttempts >= maxInteractionAttempts {
			return fmt.Errorf("failed to enter area %s after %d attempts", dest.Area().Name, maxInteractionAttempts)
		}

		// Find the entrance we want to interact with
		for _, l := range ctx.Data.AdjacentLevels {
			if l.Area == dest && l.IsEntrance {
				// Verify distance
				if ctx.PathFinder.DistanceFromMe(l.Position) > maxEntranceDistance {
					return fmt.Errorf("entrance is too far away")
				}

				// Try to interact with entrance
				if ctx.Data.HoverData.IsHovered {
					ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
				} else {
					x, y := ui.GameCoordsToScreenCords(l.Position.X-1, l.Position.Y-1)
					currentMouseCoords = data.Position{X: x, Y: y}
					ctx.HID.MovePointer(x, y)
				}

				interactionAttempts++
				utils.Sleep(200)
				break
			}
		}
	}
}
