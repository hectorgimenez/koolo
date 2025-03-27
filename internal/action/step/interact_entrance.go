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
	maxMoveRetries      = 3
)

func InteractEntrance(area area.ID) error {
	maxInteractionAttempts := 5
	interactionAttempts := 0
	waitingForInteraction := false
	currentMouseCoords := data.Position{}
	lastRun := time.Time{}

	// If we move the mouse to interact with an entrance, we will set this variable.
	var lastEntranceLevel data.Level

	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	for {
		ctx.PauseIfNotPriority()

		if ctx.Data.AreaData.Area == area && time.Since(lastRun) > time.Millisecond*500 && ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) {
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
			// It is possible to have multiple entrances to the same area (A2 sewers, A2 palace, etc)
			// Once we "select" an area and start to move the mouse to hover with it, we don't want
			// to change the area to the 2nd entrance in the same area on the next iteration.
			if l.Area == area && (lastEntranceLevel == (data.Level{}) || lastEntranceLevel == l) {
				distance := ctx.PathFinder.DistanceFromMe(l.Position)
				if distance > maxEntranceDistance {
					// Try to move closer with retries
					for retry := 0; retry < maxMoveRetries; retry++ {
						if err := MoveTo(l.Position); err != nil {
							// If MoveTo fails, try direct movement
							screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(
								l.Position.X-2,
								l.Position.Y-2,
							)
							ctx.HID.Click(game.LeftButton, screenX, screenY)
							utils.Sleep(800)
							ctx.RefreshGameData()
						}

						// Check if we're close enough now
						newDistance := ctx.PathFinder.DistanceFromMe(l.Position)
						if newDistance <= maxEntranceDistance {
							break
						}

						if retry == maxMoveRetries-1 {
							return fmt.Errorf("entrance too far away (distance: %d)", distance)
						}
					}
				}

				if l.IsEntrance {
					lx, ly := ctx.PathFinder.GameCoordsToScreenCords(l.Position.X-1, l.Position.Y-1)
					if ctx.Data.HoverData.UnitType == 5 || ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered {
						ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
						waitingForInteraction = true
						utils.Sleep(200)
					}

					x, y := utils.Spiral(interactionAttempts)
					x = x / 3
					y = y / 3
					currentMouseCoords = data.Position{X: lx + x, Y: ly + y}
					ctx.HID.MovePointer(lx+x, ly+y)
					interactionAttempts++
					utils.Sleep(100)

					lastEntranceLevel = l

					continue
				}

				return fmt.Errorf("area %s [%d] is not an entrance", area.Area().Name, area)
			}
		}
	}
}
