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
	maxMoveRetries = 3
	maxAttempts    = 15
	hoverDelay     = 100
	interactDelay  = 300
)

func InteractEntrance(targetArea area.ID) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	// Find the closest entrance(if there are 2 entrances for same destination like harem/palace cellar)
	targetLevel := findClosestEntrance(ctx, targetArea)
	if targetLevel == nil {
		return fmt.Errorf("no entrance found for area %s [%d]", targetArea.Area().Name, targetArea)
	}

	entranceDesc, found := findEntranceDescriptor(ctx, targetLevel)
	if !found {
		return fmt.Errorf("could not find entrance descriptor for area %s [%d]", targetArea.Area().Name, targetArea)
	}

	attempts := 0
	currentMouseCoords := data.Position{}
	lastAttempt := time.Time{}

	for {
		ctx.PauseIfNotPriority()

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

		if err := ensureInRange(ctx, targetLevel.Position); err != nil {
			return err
		}

		// Handle hovering and interaction .  We also need UnitType 2 here because sometimes entrances like ancient tunnel is both (unittype 2 the trap, unittype 5 to enter area)
		if ctx.Data.HoverData.UnitType == 5 || (ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered) {
			attemptInteraction(ctx, currentMouseCoords)
			attempts++
			continue
		}

		// Calculate and set new mouse position using spiral pattern
		currentMouseCoords = calculateMouseCoords(ctx, targetLevel.Position, attempts, entranceDesc)
		ctx.HID.MovePointer(currentMouseCoords.X, currentMouseCoords.Y)
		attempts++
		utils.Sleep(hoverDelay)
	}
}

func findEntranceDescriptor(ctx *context.Status, targetLevel *data.Level) (entrance.Description, bool) {
	if areaData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
		entrances := ctx.GameReader.Entrances(ctx.Data.PlayerUnit.Position, ctx.Data.HoverData)

		for _, lvl := range areaData.AdjacentLevels {
			if lvl.Position == targetLevel.Position {
				for _, e := range entrances {
					if e.Position == lvl.Position {
						return entrance.Desc[int(e.Name)], true
					}
				}
			}
		}
	}
	return entrance.Description{}, false
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

func calculateMouseCoords(ctx *context.Status, pos data.Position, attempts int, desc entrance.Description) data.Position {
	baseX, baseY := ctx.PathFinder.GameCoordsToScreenCords(pos.X, pos.Y)
	x, y := utils.EntranceSpiral(attempts, desc)
	return data.Position{X: baseX + x, Y: baseY + y}
}

func ensureInRange(ctx *context.Status, pos data.Position) error {
	for retry := 0; retry < maxMoveRetries; retry++ {
		distance := ctx.PathFinder.DistanceFromMe(pos)
		if distance <= DistanceToFinishMoving {
			return nil
		}

		if err := MoveTo(pos); err != nil {
			continue
		}

		if ctx.PathFinder.DistanceFromMe(pos) <= DistanceToFinishMoving {
			return nil
		}
		utils.Sleep(200)
	}
	return fmt.Errorf("failed to get in range of entrance after %d attempts", maxMoveRetries)
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
