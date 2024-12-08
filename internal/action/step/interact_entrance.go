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

	// Find the closest entrance
	var targetLevel *data.Level
	shortestDistance := 999999
	for _, l := range ctx.Data.AdjacentLevels {
		if l.Area == area && l.IsEntrance {
			distance := ctx.PathFinder.DistanceFromMe(l.Position)
			if distance < shortestDistance {
				shortestDistance = distance
				lvl := l // Create a new variable to avoid pointer issues
				targetLevel = &lvl
			}
		}
	}

	if targetLevel == nil {
		return fmt.Errorf("no entrance found for area %s [%d]", area.Area().Name, area)
	}

	for {
		ctx.PauseIfNotPriority()

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

		distance := ctx.PathFinder.DistanceFromMe(targetLevel.Position)
		if distance > maxEntranceDistance {
			for retry := 0; retry < maxMoveRetries; retry++ {
				if ctx.Data.OpenMenus.LoadingScreen {
					ctx.Logger.Debug("Loading screen detected during movement...")
					ctx.WaitForGameToLoad()
					break
				}

				if err := MoveTo(targetLevel.Position); err != nil {
					screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(
						targetLevel.Position.X-2,
						targetLevel.Position.Y-2,
					)
					ctx.HID.Click(game.LeftButton, screenX, screenY)
					utils.Sleep(800)
					ctx.RefreshGameData()
				}

				newDistance := ctx.PathFinder.DistanceFromMe(targetLevel.Position)
				if newDistance <= maxEntranceDistance {
					break
				}

				if retry == maxMoveRetries-1 {
					return fmt.Errorf("entrance too far away (distance: %d)", distance)
				}
			}
		}

		// Store successful hover position
		if ctx.Data.HoverData.UnitType == 5 || ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered {
			lastHoverCoords = currentMouseCoords
		}

		if ctx.Data.HoverData.UnitType == 5 || ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			waitingForInteraction = true

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

		lx, ly := ctx.PathFinder.GameCoordsToScreenCords(targetLevel.Position.X-1, targetLevel.Position.Y-1)
		x, y := utils.Spiral(interactionAttempts)
		x = x / 3
		y = y / 3
		currentMouseCoords = data.Position{X: lx + x, Y: ly + y}
		ctx.HID.MovePointer(lx+x, ly+y)
		interactionAttempts++
		utils.Sleep(100)

		if ctx.Data.AreaData.Area == area && ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) {
			if ctx.Data.OpenMenus.LoadingScreen {
				ctx.WaitForGameToLoad()
			}

			ctx.CurrentGame.AreaCorrection.Enabled = true
			ctx.CurrentGame.AreaCorrection.ExpectedArea = area
			return nil
		}
	}
}
