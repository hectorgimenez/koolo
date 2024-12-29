package step

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/entrance"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxMoveRetries = 6
	maxAttempts    = 20
	hoverDelay     = 50
	interactDelay  = 100
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
	lastAttempt := time.Now()

	for {
		ctx.PauseIfNotPriority()
		ctx.RefreshGameData()

		// Handle loading screen early
		if ctx.Data.OpenMenus.LoadingScreen {
			ctx.WaitForGameToLoad()
			continue
		}

		// Check if we've reached the target area
		if ctx.Data.AreaData.Area == targetArea &&
			time.Since(lastAttempt) > interactDelay &&
			ctx.Data.AreaData.IsInside(ctx.Data.PlayerUnit.Position) {
			return nil
		}

		if attempts >= maxAttempts {
			return fmt.Errorf("failed to enter area %s after %d attempts", targetArea.Area().Name, maxAttempts)
		}

		if time.Since(lastAttempt) < interactDelay {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Ensure we're in range of the entrance
		if err := ensureInRange(ctx, targetLevel.Position); err != nil {
			return err
		}

		// Handle hovering and interaction .  We also need UnitType 2 here because sometimes entrances like ancient tunnel is both (unittype 2 the trap, unittype 5 to enter area)
		if ctx.Data.HoverData.UnitType == 5 || (ctx.Data.HoverData.UnitType == 2 && ctx.Data.HoverData.IsHovered) {
			ctx.HID.Click(game.LeftButton, currentMouseCoords.X, currentMouseCoords.Y)
			lastAttempt = time.Now()
			attempts++

			// Wait for loading screen after click
			startTime := time.Now()
			for time.Since(startTime) < time.Second {
				if ctx.Data.OpenMenus.LoadingScreen {
					ctx.WaitForGameToLoad()
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
			continue
		}

		// Calculate new mouse position using spiral pattern
		currentMouseCoords = calculateMouseCoords(ctx, targetLevel.Position, attempts, entranceDesc)
		ctx.HID.MovePointer(currentMouseCoords.X, currentMouseCoords.Y)
		attempts++

		time.Sleep(hoverDelay * time.Millisecond)
	}
}

func findEntranceDescriptor(ctx *context.Status, targetLevel *data.Level) (entrance.Description, bool) {
	maxRetries := 5
	baseDelay := 500 * time.Millisecond

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			sleepTime := time.Duration(retry)*baseDelay + time.Duration(rand.Intn(200))*time.Millisecond
			time.Sleep(sleepTime)
			ctx.RefreshGameData()
			ctx.Logger.Debug("Retrying to find entrance descriptor",
				"attempt", retry+1,
				"area", targetLevel.Area.Area().Name,
				"sleep", sleepTime.String())
		}

		if areaData, ok := ctx.Data.Areas[ctx.Data.PlayerUnit.Area]; ok {
			entrances := ctx.GameReader.Entrances(ctx.Data.PlayerUnit.Position, ctx.Data.HoverData)

			// First check if we have any entrances at all
			if len(entrances) == 0 {
				continue
			}

			// Then look for our specific entrance
			for _, lvl := range areaData.AdjacentLevels {
				if lvl.Position == targetLevel.Position {
					for _, e := range entrances {
						if e.Position == lvl.Position {
							desc, exists := entrance.Desc[int(e.Name)]
							if exists {
								return desc, true
							}
						}
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

func ensureInRange(ctx *context.Status, pos data.Position) error {
	requiredDistance := 3 // We need to be at distance 3 or less to interact with entrances

	for retry := 0; retry < maxMoveRetries; retry++ {
		distance := ctx.PathFinder.DistanceFromMe(pos)
		if distance <= requiredDistance {
			return nil
		}

		var moveOpts []MoveOption
		// Gradually reduce target distance on each attempt
		if retry >= 4 {
			moveOpts = append(moveOpts, WithDistanceToFinish(1))
		} else if retry >= 2 {
			moveOpts = append(moveOpts, WithDistanceToFinish(2))
		} else if distance == 4 {
			moveOpts = append(moveOpts, WithDistanceToFinish(3))
		}

		if err := MoveTo(pos, moveOpts...); err != nil {
			// On error, if we've tried a few times, check for and kill blocking monsters
			if retry >= 2 {
				entrancePos := pos // Capture for closure
				ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
					var closestMonster data.Monster
					shortestDistance := 999999
					found := false

					for _, m := range d.Monsters {
						// Calculate distance from monster to entrance
						dx := float64(m.Position.X - entrancePos.X)
						dy := float64(m.Position.Y - entrancePos.Y)
						monsterToEntranceDist := int(math.Sqrt(dx*dx + dy*dy))

						if monsterToEntranceDist <= 5 {
							dist := ctx.PathFinder.DistanceFromMe(m.Position)
							if dist < shortestDistance {
								shortestDistance = dist
								closestMonster = m
								found = true
							}
						}
					}
					return closestMonster.UnitID, found
				}, nil)
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		// Check if we succeeded at getting to required distance
		if ctx.PathFinder.DistanceFromMe(pos) <= requiredDistance {
			return nil
		}
	}

	return fmt.Errorf("failed to get in range of entrance after %d attempts", maxMoveRetries)
}

func calculateMouseCoords(ctx *context.Status, pos data.Position, attempts int, desc entrance.Description) data.Position {
	baseX, baseY := ctx.PathFinder.GameCoordsToScreenCords(pos.X, pos.Y)
	x, y := utils.EntranceSpiral(attempts, desc)
	return data.Position{X: baseX + x, Y: baseY + y}
}
