package action

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/utils"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	maxAreaSyncAttempts       = 10
	areaSyncDelay             = 100 * time.Millisecond
	monsterCacheTime          = 30 * time.Second
	monsterHandleCooldown     = 1500 * time.Millisecond
	monsterMapCleanupInterval = 5 * time.Minute
)

// Cache structure to avoid frequent CPU intensive monster checks
var actionAttemptedMonsterClears = make(map[data.UnitID]time.Time)
var actionLastMonsterHandlingTime = time.Time{}
var actionLastMonsterMapCleanup = time.Now()

func ensureAreaSync(ctx *context.Status, expectedArea area.ID) error {
	// Skip sync check if we're already in the expected area and have valid area data
	if ctx.Data.PlayerUnit.Area == expectedArea {
		if areaData, ok := ctx.Data.Areas[expectedArea]; ok && areaData.IsInside(ctx.Data.PlayerUnit.Position) {
			return nil
		}
	}

	// Wait for area data to sync
	for attempts := 0; attempts < maxAreaSyncAttempts; attempts++ {
		ctx.RefreshGameData()

		if ctx.Data.PlayerUnit.Area == expectedArea {
			if areaData, ok := ctx.Data.Areas[expectedArea]; ok {
				if areaData.IsInside(ctx.Data.PlayerUnit.Position) {
					return nil
				}
			}
		}

		time.Sleep(areaSyncDelay)
	}

	return fmt.Errorf("area sync timeout - expected: %v, current: %v", expectedArea, ctx.Data.PlayerUnit.Area)
}

func MoveToArea(dst area.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("MoveToArea")

	if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		return err
	}

	// Exceptions for:
	// Arcane Sanctuary
	if dst == area.ArcaneSanctuary && ctx.Data.PlayerUnit.Area == area.PalaceCellarLevel3 {
		ctx.Logger.Debug("Arcane Sanctuary detected, finding the Portal")
		portal, _ := ctx.Data.Objects.FindOne(object.ArcaneSanctuaryPortal)
		MoveToCoords(portal.Position)

		return step.InteractObject(portal, func() bool {
			return ctx.Data.PlayerUnit.Area == area.ArcaneSanctuary
		})
	}
	// Canyon of the Magi
	if dst == area.CanyonOfTheMagi && ctx.Data.PlayerUnit.Area == area.ArcaneSanctuary {
		ctx.Logger.Debug("Canyon of the Magi detected, finding the Portal")
		tome, _ := ctx.Data.Objects.FindOne(object.YetAnotherTome)
		MoveToCoords(tome.Position)
		InteractObject(tome, func() bool {
			if _, found := ctx.Data.Objects.FindOne(object.PermanentTownPortal); found {
				ctx.Logger.Debug("Opening YetAnotherTome!")
				return true
			}
			return false
		})
		ctx.Logger.Debug("Using Canyon of the Magi Portal")
		portal, _ := ctx.Data.Objects.FindOne(object.PermanentTownPortal)
		MoveToCoords(portal.Position)
		return step.InteractObject(portal, func() bool {
			return ctx.Data.PlayerUnit.Area == area.CanyonOfTheMagi
		})
	}

	lvl := data.Level{}
	for _, a := range ctx.Data.AdjacentLevels {
		if a.Area == dst {
			lvl = a
			break
		}
	}

	if lvl.Position.X == 0 && lvl.Position.Y == 0 {
		return fmt.Errorf("destination area not found: %s", dst.Area().Name)
	}

	toFun := func() (data.Position, bool) {
		if ctx.Data.PlayerUnit.Area == dst {
			ctx.Logger.Debug("Reached area", slog.String("area", dst.Area().Name))
			return data.Position{}, false
		}

		if ctx.Data.PlayerUnit.Area == area.TamoeHighland && dst == area.MonasteryGate {
			ctx.Logger.Debug("Monastery Gate detected, moving to static coords")
			return data.Position{X: 15139, Y: 5056}, true
		}

		if ctx.Data.PlayerUnit.Area == area.MonasteryGate && dst == area.TamoeHighland {
			ctx.Logger.Debug("Monastery Gate detected, moving to static coords")
			return data.Position{X: 15142, Y: 5118}, true
		}

		// To correctly detect the two possible exits from Lut Gholein
		if dst == area.RockyWaste && ctx.Data.PlayerUnit.Area == area.LutGholein {
			if _, _, found := ctx.PathFinder.GetPath(data.Position{X: 5004, Y: 5065}); found {
				return data.Position{X: 4989, Y: 5063}, true
			} else {
				return data.Position{X: 5096, Y: 4997}, true
			}
		}

		// This means it's a cave, we don't want to load the map, just find the entrance and interact
		if lvl.IsEntrance {
			return lvl.Position, true
		}

		objects := ctx.Data.Areas[lvl.Area].Objects
		// Sort objects by the distance from me
		sort.Slice(objects, func(i, j int) bool {
			distanceI := ctx.PathFinder.DistanceFromMe(objects[i].Position)
			distanceJ := ctx.PathFinder.DistanceFromMe(objects[j].Position)

			return distanceI < distanceJ
		})

		// Let's try to find any random object to use as a destination point, once we enter the level we will exit this flow
		for _, obj := range objects {
			_, _, found := ctx.PathFinder.GetPath(obj.Position)
			if found {
				return obj.Position, true
			}
		}

		return lvl.Position, true
	}

	err := MoveTo(toFun)
	if err != nil {
		ctx.Logger.Warn("error moving to area, will try to continue", slog.String("error", err.Error()))
	}

	if lvl.IsEntrance {
		maxAttempts := 3
		for attempt := 0; attempt < maxAttempts; attempt++ {
			// Check current distance
			currentDistance := ctx.PathFinder.DistanceFromMe(lvl.Position)

			if currentDistance > 7 {
				// For distances > 7, recursively call MoveToArea as it includes the entrance interaction
				return MoveToArea(dst)
			} else if currentDistance > 3 && currentDistance <= 7 {
				// For distances between 4 and 7, use direct click
				screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(
					lvl.Position.X-2,
					lvl.Position.Y-2,
				)
				ctx.HID.Click(game.LeftButton, screenX, screenY)
				utils.Sleep(800)
			}

			// Try to interact with the entrance
			err = step.InteractEntrance(dst)
			if err == nil {
				break
			}

			if attempt < maxAttempts-1 {
				ctx.Logger.Debug("Entrance interaction failed, retrying",
					slog.Int("attempt", attempt+1),
					slog.String("error", err.Error()))
				utils.Sleep(1000)
			}
		}

		if err != nil {
			return fmt.Errorf("failed to interact with area %s after %d attempts: %v", dst.Area().Name, maxAttempts, err)
		}

		// Wait for area transition to complete
		if err := ensureAreaSync(ctx, dst); err != nil {
			return err
		}
	}

	event.Send(event.InteractedTo(event.Text(ctx.Name, ""), int(dst), event.InteractionTypeEntrance))
	return nil
}

func MoveToCoords(to data.Position) error {
	ctx := context.Get()

	if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		return err
	}

	return MoveTo(func() (data.Position, bool) {
		return to, true
	})
}

func MoveTo(toFunc func() (data.Position, bool)) error {
	ctx := context.Get()
	ctx.SetLastAction("MoveTo")

	// Ensure no menus are open that might block movement
	for ctx.Data.OpenMenus.IsMenuOpen() {
		ctx.Logger.Debug("Found open menus while moving, closing them...")
		if err := step.CloseAllMenus(); err != nil {
			return err
		}

		utils.Sleep(500)
	}

	lastMovement := false

	// Initial sync check
	if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		return err
	}

	// Clean up monster map occasionally
	if time.Since(actionLastMonsterMapCleanup) > monsterMapCleanupInterval {
		cleanupMonsterMap()
		actionLastMonsterMapCleanup = time.Now()
	}

	for {
		ctx.RefreshGameData()
		to, found := toFunc()
		if !found {
			return nil
		}

		// If we can teleport, don't bother with the rest
		if ctx.Data.CanTeleport() {
			return step.MoveTo(to)
		}

		if lastMovement {
			return nil
		}

		// TODO: refactor this to use the same approach as ClearThroughPath
		if _, distance, _ := ctx.PathFinder.GetPathFrom(ctx.Data.PlayerUnit.Position, to); distance <= step.DistanceToFinishMoving {
			lastMovement = true
		}

		moveErr := step.MoveTo(to)
		if moveErr != nil {
			// Handle the monsters in path error
			if errors.Is(moveErr, step.ErrMonstersInPath) {
				if time.Since(actionLastMonsterHandlingTime) < monsterHandleCooldown {
					// Skip handling, just continue
					time.Sleep(100 * time.Millisecond)
					continue
				}

				actionLastMonsterHandlingTime = time.Now()
				ctx.Logger.Debug("Monsters detected in path, clearing before continuing")

				// Fast check if we've already attempted to clear any monsters in this area recently
				skipClearing := false
				clearPathDist := ctx.CharacterCfg.Character.ClearPathDist

				// Get list of nearby monsters more efficiently
				nearbyMonsters := make([]data.UnitID, 0, 5) // Pre-allocate small capacity
				for _, m := range ctx.Data.Monsters.Enemies() {
					if m.Stats[stat.Life] <= 0 {
						continue
					}

					distanceToMonster := ctx.PathFinder.DistanceFromMe(m.Position)
					if distanceToMonster <= clearPathDist {
						nearbyMonsters = append(nearbyMonsters, m.UnitID)

						// Check if we recently tried to clear this monster - early exit optimization
						if clearTime, exists := actionAttemptedMonsterClears[m.UnitID]; exists {
							if time.Since(clearTime) < monsterCacheTime {
								skipClearing = true
								break // Early exit if we find any monster we've already tried to clear
							}
						}
					}
				}

				if skipClearing {
					ctx.Logger.Debug("Skipping monster clearing - already attempted recently")
				} else if len(nearbyMonsters) > 0 {
					// Try clearing monsters
					clearErr := ClearAreaAroundPosition(ctx.Data.PlayerUnit.Position, clearPathDist, data.MonsterAnyFilter())

					// Mark all these monsters as attempted, regardless of success
					now := time.Now()
					for _, monsterID := range nearbyMonsters {
						actionAttemptedMonsterClears[monsterID] = now
					}

					if clearErr != nil {
						ctx.Logger.Warn("Failed to clear all monsters, will try to path around them",
							slog.String("error", fmt.Sprintf("%v", clearErr)))
					}
				}

				// Continue the path
				continue
			}

			return moveErr
		}
	}
}

func cleanupMonsterMap() {
	if len(actionAttemptedMonsterClears) > 100 { // Only clean if the map is getting large
		now := time.Now()
		for id, timestamp := range actionAttemptedMonsterClears {
			if now.Sub(timestamp) > monsterMapCleanupInterval {
				delete(actionAttemptedMonsterClears, id)
			}
		}
	}
}
