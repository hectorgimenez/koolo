package step

import (
	"fmt"
	"math"
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
	maxAttempts         = 5
	hoverDelay          = 100
	interactDelay       = 500
)

func InteractEntrance(area area.ID) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	// Find the closest entrance(if there is 2 entrance for same destination like harem/palace cellar)
	targetLevel := findClosestEntrance(ctx, area)
	if targetLevel == nil {
		return fmt.Errorf("no entrance found for area %s [%d]", area.Area().Name, area)
	}

	attempts := 0
	currentMouseCoords := data.Position{}
	lastHoverCoords := data.Position{}
	lastAttempt := time.Time{}
	useOriginalSpiral := false

	// Find the entrance object at the target position
	var entranceObj data.Object
	var foundEntranceObj bool
	for _, obj := range ctx.Data.Objects {
		if obj.Position == targetLevel.Position {
			entranceObj = obj
			foundEntranceObj = true
			break
		}
	}

	// If we can't find the entrance object, we'll only use the original spiral
	if !foundEntranceObj {
		useOriginalSpiral = true
	}

	for {
		ctx.PauseIfNotPriority()

		// Handle loading screens
		if ctx.Data.OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			continue
		}

		if hasReachedArea(ctx, area, lastAttempt) {
			ctx.CurrentGame.AreaCorrection.Enabled = true
			ctx.CurrentGame.AreaCorrection.ExpectedArea = area
			return nil
		}

		if attempts >= maxAttempts {
			if !useOriginalSpiral {
				// Switch to original spiral pattern and reset attempts
				useOriginalSpiral = true
				attempts = 0
				lastHoverCoords = data.Position{}
				continue
			}
			return fmt.Errorf("failed to enter area %s after all attempts", area.Area().Name)
		}

		if time.Since(lastAttempt) < interactDelay {
			continue
		}
		lastAttempt = time.Now()

		// Move closer if needed
		if err := ensureInRange(ctx, targetLevel.Position); err != nil {
			return err
		}

		// Handle hovering and interaction
		if ctx.Data.HoverData.UnitType == 5 || (ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered) {
			lastHoverCoords = currentMouseCoords
			attemptInteraction(ctx, currentMouseCoords)
			attempts++
			continue
		}

		// Try last successful hover position first(if we hovered it once then we have exact interaction point)
		if attempts == 0 && lastHoverCoords != (data.Position{}) {
			currentMouseCoords = lastHoverCoords
			ctx.HID.MovePointer(lastHoverCoords.X, lastHoverCoords.Y)
			attempts++
			utils.Sleep(hoverDelay)
			continue
		}

		baseX, baseY := ctx.PathFinder.GameCoordsToScreenCords(targetLevel.Position.X-2, targetLevel.Position.Y-2)
		var x, y int
		if useOriginalSpiral {
			x, y = utils.Spiral(attempts)
			x = x / 3
			y = y / 3
		} else {
			desc := entranceObj.Desc()
			x, y = entranceSpiral(attempts, desc.SizeX, desc.SizeY)
		}

		currentMouseCoords = data.Position{X: baseX + x, Y: baseY + y}
		ctx.HID.MovePointer(currentMouseCoords.X, currentMouseCoords.Y)
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

		screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(pos.X-2, pos.Y-2)
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
func entranceSpiral(attempt int, sizeX, sizeY int) (x, y int) {
	// Use golden ratio angle distribution for all cases
	baseRadius := float64(attempt) * 3.0
	angle := float64(attempt) * math.Pi * (3.0 - math.Sqrt(5.0))

	// If entrance has no dimensions (size 0), use default entrance pattern
	if sizeX == 0 || sizeY == 0 {
		xScale := 1.2 // Wider search pattern
		yScale := 0.8 // Shorter vertical search

		x = int(baseRadius * math.Cos(angle) * xScale)
		y = int(baseRadius * math.Sin(angle) * yScale)

		// Add a slight upward bias for entrances
		y -= 35

		// Use fixed bounds for zero-dimension entrances
		x = utils.Clamp(x, -30, 30)
		y = utils.Clamp(y, -50, 10)

		return x, y
	}

	// For entrances with dimensions, use size-based scaling
	xScale := 1.2
	yScale := 0.8
	if sizeY > sizeX {
		// Adjust scales for vertical entrances
		xScale = 1.4
		yScale = 0.7
	}

	x = int(baseRadius * math.Cos(angle) * xScale)
	y = int(baseRadius * math.Sin(angle) * yScale)
	y -= 40

	x = utils.Clamp(x, -35, 35)
	y = utils.Clamp(y, -60, 0)

	return x, y
}
