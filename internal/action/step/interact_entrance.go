package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/entrance"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxEntranceDistance = 6
	maxMoveRetries      = 3
	maxAttempts         = 10
	hoverDelay          = 100
	interactDelay       = 250
)

func InteractEntrance(targetArea area.ID) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	// Find the closest entrance(if there are 2 entrances for same destination like harem/palace cellar)
	targetLevel := findClosestEntrance(ctx, targetArea)
	if targetLevel == nil {
		return fmt.Errorf("no entrance found for area %s [%d]", targetArea.Area().Name, targetArea)
	}

	// Find entrance descriptor through 1.13c map data
	var entranceDesc entrance.Description
	var descFound bool

	if areaData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
		// Get entrances for current area
		entrances := ctx.GameReader.Entrances(ctx.Data.PlayerUnit.Position, ctx.Data.HoverData)

		// For each adjacent level that matches our target
		for _, lvl := range areaData.AdjacentLevels {
			if lvl.Position == targetLevel.Position {
				// Find matching entrance by position
				for _, e := range entrances {
					if e.Position == lvl.Position {
						// Use the entrance Name to get the proper description
						entranceDesc = entrance.Desc[int(e.Name)]
						descFound = true
						break
					}
				}
				break
			}
		}
	}

	if !descFound {
		return fmt.Errorf("could not find entrance descriptor for area %s [%d]", targetArea.Area().Name, targetArea)
	}

	attempts := 0
	currentMouseCoords := data.Position{}
	lastAttempt := time.Time{}

	for {
		ctx.PauseIfNotPriority()

		// Handle loading screens
		if ctx.Data.OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			continue
		}

		if hasReachedArea(ctx, targetArea, lastAttempt) {
			return nil
		}

		if attempts >= maxAttempts {
			return fmt.Errorf("failed to enter area %s after all attempts", targetArea.Area().Name)
		}

		if time.Since(lastAttempt) < interactDelay {
			continue
		}
		lastAttempt = time.Now()

		// Move closer if needed
		if err := ensureInRange(ctx, targetLevel.Position); err != nil {
			return err
		}

		// Handle hovering and interaction .  We also need UnitType 2 here because sometimes entrances like ancient tunnel is both (unittype 2 the trap, unittype 5 to enter area)
		if ctx.Data.HoverData.UnitType == 5 || (ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered) {
			attemptInteraction(ctx, currentMouseCoords)
			attempts++
			continue
		}

		// Calculate the mouse position for interaction
		baseX, baseY := ctx.PathFinder.GameCoordsToScreenCords(targetLevel.Position.X, targetLevel.Position.Y)
		x, y := utils.EntranceSpiral(attempts, entranceDesc)
		currentMouseCoords = data.Position{X: baseX + x, Y: baseY + y}
		ctx.HID.MovePointer(currentMouseCoords.X, currentMouseCoords.Y)

		// Increment attempt count and wait before retrying
		attempts++
		utils.Sleep(hoverDelay)
	}
}

func findClosestEntrance(ctx *context.Status, targetArea area.ID) *data.Level {
	var closest *data.Level
	shortestDistance := 999999
	for _, l := range ctx.Data.AdjacentLevels {
		if l.Area == targetArea && l.IsEntrance {
			distance := ctx.PathFinder.DistanceFromMe(l.Position)
			if distance < shortestDistance {
				shortestDistance = distance
				lvl := l
				closest = &lvl
			}
		}
	}
	return closest
}

func hasReachedArea(ctx *context.Status, targetArea area.ID, lastAttempt time.Time) bool {
	return ctx.Data.AreaData.Area == targetArea &&
		time.Since(lastAttempt) > interactDelay &&
		ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position)
}

func ensureInRange(ctx *context.Status, pos data.Position) error {
	distance := ctx.PathFinder.DistanceFromMe(pos)
	if distance <= maxEntranceDistance {
		return nil
	}

	// Direct MoveTo for longer distances
	if distance >= 7 {
		return MoveTo(pos)
	}

	// For shorter distances, try clicking
	for retry := 0; retry < maxMoveRetries; retry++ {
		if ctx.Data.OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			break
		}

		screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(pos.X, pos.Y)
		ctx.HID.Click(game.LeftButton, screenX, screenY)
		utils.Sleep(800)
		ctx.RefreshGameData()

		if ctx.PathFinder.DistanceFromMe(pos) <= maxEntranceDistance {
			return nil
		}
	}

	return fmt.Errorf("failed to get in range of entrance (distance: %d)", distance)
}

func attemptInteraction(ctx *context.Status, pos data.Position) {
	ctx.HID.Click(game.LeftButton, pos.X, pos.Y)
	startTime := time.Now()
	for time.Since(startTime) < 2*time.Second {
		if ctx.Data.OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			break
		}
		utils.Sleep(50)
	}
	utils.Sleep(200)
}
