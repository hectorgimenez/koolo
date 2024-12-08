package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
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
	lastHoverCoords := data.Position{}
	lastRun := time.Time{}

	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	for {
		ctx.PauseIfNotPriority()

		// Check if we're in loading screen
		if ctx.Data.OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			continue
		}

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
			if l.Area == area {
				distance := ctx.PathFinder.DistanceFromMe(l.Position)
				if distance > maxEntranceDistance {
					// Try to move closer with retries
					for retry := 0; retry < maxMoveRetries; retry++ {
						// Check for loading screen during movement
						if ctx.Data.OpenMenus.LoadingScreen {
							ctx.Logger.Debug("Loading screen detected during movement...")
							ctx.WaitForGameToLoad()
							break
						}

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
					// Store successful hover position
					if ctx.Data.HoverData.UnitType == 5 || ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered {
						lastHoverCoords = currentMouseCoords
					}

					if ctx.Data.HoverData.UnitType == 5 || ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered {
						ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
						waitingForInteraction = true

						// Wait for loading screen after clicking entrance
						startTime := time.Now()
						for time.Since(startTime) < time.Second*2 {
							if ctx.Data.OpenMenus.LoadingScreen {
								ctx.WaitForGameToLoad()
								break
							}
							utils.Sleep(50)
						}

						utils.Sleep(200)
					}

					// Try last successful hover position first
					if interactionAttempts == 0 && lastHoverCoords != (data.Position{}) {
						currentMouseCoords = lastHoverCoords
						ctx.HID.MovePointer(lastHoverCoords.X, lastHoverCoords.Y)
						interactionAttempts++
						utils.Sleep(100)
						continue
					}

					lx, ly := ctx.PathFinder.GameCoordsToScreenCords(l.Position.X-1, l.Position.Y-1)
					x, y := utils.Spiral(interactionAttempts)
					x = x / 3 // Tighter spiral for better precision
					y = y / 3
					currentMouseCoords = data.Position{X: lx + x, Y: ly + y}
					ctx.HID.MovePointer(lx+x, ly+y)
					interactionAttempts++
					utils.Sleep(100)
					continue
				}

				// After successful entrance transition and area sync
				if ctx.Data.AreaData.Area == area && ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) {
					// Wait for loading to complete before enabling area correction
					if ctx.Data.OpenMenus.LoadingScreen {
						ctx.WaitForGameToLoad()
					}

					// Enable area correction with this area as expected
					ctx.CurrentGame.AreaCorrection.Enabled = true
					ctx.CurrentGame.AreaCorrection.ExpectedArea = area
					return nil
				}

				return fmt.Errorf("area %s [%d] is not an entrance", area.Area().Name, area)
			}
		}
	}
}
