package step

import (
	"fmt"
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/entrance"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxAttempts      = 30
	requiredDistance = 3
)

func InteractEntrance(targetArea area.ID) error {
	ctx := context.Get()
	ctx.SetLastStep("InteractEntrance")

	// Find the closest entrance(if there are 2 entrances for same destination like harem/palace cellar)
	targetLevel := findClosestEntrance(ctx, targetArea)
	if targetLevel == nil {
		return fmt.Errorf("no entrance found for area %s [%d]", targetArea.Area().Name, targetArea)
	}
	// link adjacentlvl to entrance unit and return its description ( entrance.desc )
	desc, found := findEntranceDescriptor(ctx, targetLevel)
	if !found {
		return fmt.Errorf("could not find entrance descriptor for area %s", targetArea.Area().Name)
	}

	attempts := 0
	lastAttempt := time.Now()
	currentMousePos := data.Position{}
	hasTriedClick := false

	for {
		ctx.PauseIfNotPriority()
		ctx.RefreshGameData()

		// Handle loading screen early
		if ctx.Data.OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			continue
		}

		// Check if we've reached target area
		if ctx.Data.AreaData.Area == targetArea && ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) {
			return nil
		}

		if attempts >= maxAttempts {
			return fmt.Errorf("failed to enter area %s after %d attempts", targetArea.Area().Name, maxAttempts)
		}

		if time.Since(lastAttempt) < 100*time.Millisecond {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Move to entrance if needed
		if err := moveToEntrance(ctx, targetLevel.Position); err != nil {
			return err
		}

		// If we're in range but can't interact, try clicking once to get closer.
		// Sometimes pathfinder end up at an angle where it cant hover the entrance
		dist := ctx.PathFinder.DistanceFromMe(targetLevel.Position)
		if !hasTriedClick && dist <= 4 && attempts > 10 {
			baseX, baseY := ctx.PathFinder.GameCoordsToScreenCords(targetLevel.Position.X-2, targetLevel.Position.Y-2)
			ctx.HID.Click(game.LeftButton, baseX, baseY)
			hasTriedClick = true
			time.Sleep(800 * time.Millisecond)
			continue
		}

		// Handle hovering and interaction. We also need UnitType 2 here because sometimes entrances like ancient tunnel is both (unittype 2 the trap, unittype 5 to enter area)
		if ctx.Data.HoverData.UnitType == 5 || (ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered) {
			ctx.HID.Click(game.LeftButton, currentMousePos.X, currentMousePos.Y)
			lastAttempt = time.Now()
			attempts++
			continue
		}

		// Try new hover position
		baseX, baseY := ctx.PathFinder.GameCoordsToScreenCords(targetLevel.Position.X, targetLevel.Position.Y)
		offsetX, offsetY := utils.EntranceSpiral(attempts, desc)
		currentMousePos = data.Position{X: baseX + offsetX, Y: baseY + offsetY}
		ctx.HID.MovePointer(currentMousePos.X, currentMousePos.Y)
		attempts++
		time.Sleep(50 * time.Millisecond)
	}
}
func moveToEntrance(ctx *context.Status, pos data.Position) error {
	distance := ctx.PathFinder.DistanceFromMe(pos)
	if distance <= requiredDistance {
		return nil
	}

	var moveOpts []MoveOption
	moveOpts = append(moveOpts, WithDistanceToFinish(2))

	err := MoveTo(pos, moveOpts...)
	if err != nil {
		return err
	}
	return nil
}

func findClosestEntrance(ctx *context.Status, targetArea area.ID) *data.Level {
	var closest *data.Level
	shortestDist := math.MaxInt32

	for _, l := range ctx.Data.AdjacentLevels {
		if l.Area == targetArea && l.IsEntrance {
			dist := ctx.PathFinder.DistanceFromMe(l.Position)
			if dist < shortestDist {
				shortestDist = dist
				lvl := l
				closest = &lvl
			}
		}
	}
	return closest
}

func findEntranceDescriptor(ctx *context.Status, targetLevel *data.Level) (entrance.Description, bool) {
	if areaData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
		entrances := ctx.GameReader.Entrances(ctx.Data.PlayerUnit.Position, ctx.Data.HoverData)
		for _, lvl := range areaData.AdjacentLevels {
			if lvl.Position == targetLevel.Position {
				for _, e := range entrances {
					if e.Position == lvl.Position {
						desc, exists := entrance.Desc[int(e.Name)]
						return desc, exists
					}
				}
			}
		}
	}
	return entrance.Description{}, false
}
