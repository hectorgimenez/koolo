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
	currentMousePos := data.Position{}
	hasTriedClick := false

	// Initial state setup
	lastPlayerPos := ctx.Data.PlayerUnit.Position
	lastArea := ctx.Data.PlayerUnit.Area
	lastAreaData := ctx.Data.AreaData

	for {
		ctx.PauseIfNotPriority()

		// Handle loading screen early
		if ctx.GameReader.GetData().OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			continue
		}

		// Get minimal required state updates
		gr := ctx.GameReader.GameReader
		hover := gr.GetData().HoverData

		// Check current area - only update if loading screen was seen
		if gr.InGame() {
			rawPlayerUnits := gr.GetRawPlayerUnits()
			mainPlayer := rawPlayerUnits.GetMainPlayer()
			if mainPlayer.Area != lastArea {
				// Full refresh needed for area transition
				ctx.RefreshGameData()
				lastArea = ctx.Data.PlayerUnit.Area
				lastAreaData = ctx.Data.AreaData
				lastPlayerPos = ctx.Data.PlayerUnit.Position
			}
		}

		// Check if we've reached target area
		if lastArea == targetArea && lastAreaData.IsInside(lastPlayerPos) {
			return nil
		}

		if attempts >= maxAttempts {
			return fmt.Errorf("failed to enter area %s after %d attempts", targetArea.Area().Name, maxAttempts)
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
		if hover.UnitType == 5 || (hover.UnitType == 2 && hover.IsHovered) {
			ctx.HID.Click(game.LeftButton, currentMousePos.X, currentMousePos.Y)
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
