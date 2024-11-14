package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	maxAreaSyncAttempts = 10
	areaSyncDelay       = 100 * time.Millisecond
)

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
	ctx.CurrentGame.AreaCorrection.Enabled = false
	defer func() {
		ctx.CurrentGame.AreaCorrection.ExpectedArea = dst
		ctx.CurrentGame.AreaCorrection.Enabled = true
	}()

	if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		return err
	}

	// Exception for Arcane Sanctuary
	if dst == area.ArcaneSanctuary && ctx.Data.PlayerUnit.Area == area.PalaceCellarLevel3 {
		ctx.Logger.Debug("Arcane Sanctuary detected, finding the Portal")
		portal, _ := ctx.Data.Objects.FindOne(object.ArcaneSanctuaryPortal)
		MoveToCoords(portal.Position)

		return step.InteractObject(portal, func() bool {
			return ctx.Data.PlayerUnit.Area == area.ArcaneSanctuary
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
			// If we're too far, try to get closer using direct clicks
			if currentDistance > 4 {
				ctx.Logger.Debug("Attempting to move closer to entrance using direct movement")

				// Calculate screen coordinates for a position closer to the entrance
				screenX, screenY := ctx.PathFinder.GameCoordsToScreenCords(
					lvl.Position.X-2,
					lvl.Position.Y-2,
				)

				// Use direct click to move closer
				ctx.HID.Click(game.LeftButton, screenX, screenY)

				// Give time for movement
				utils.Sleep(800)
				ctx.RefreshGameData()

				// Verify new position
				newDistance := ctx.PathFinder.DistanceFromMe(lvl.Position)
				ctx.Logger.Debug("New distance after move attempt",
					slog.Int("distance", newDistance),
					slog.Int("attempt", attempt+1))
			}

			err = step.InteractEntrance(dst)
			if err == nil {
				break
			}

			if attempt < maxAttempts-1 {
				ctx.Logger.Debug("Entrance interaction failed, retrying...",
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
	ctx.CurrentGame.AreaCorrection.Enabled = false
	defer func() {
		ctx.CurrentGame.AreaCorrection.ExpectedArea = ctx.Data.AreaData.Area
		ctx.CurrentGame.AreaCorrection.Enabled = true
	}()

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

	openedDoors := make(map[object.Name]data.Position)
	previousIterationPosition := data.Position{}
	lastMovement := false

	// Initial sync check
	if err := ensureAreaSync(ctx, ctx.Data.PlayerUnit.Area); err != nil {
		return err
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

		// Check for doors blocking path
		for _, o := range ctx.Data.Objects {
			if o.IsDoor() && ctx.PathFinder.DistanceFromMe(o.Position) < 10 && openedDoors[o.Name] != o.Position {
				if o.Selectable {
					ctx.Logger.Info("Door detected and teleport is not available, trying to open it...")
					openedDoors[o.Name] = o.Position
					err := step.InteractObject(o, func() bool {
						obj, found := ctx.Data.Objects.FindByID(o.ID)
						return found && !obj.Selectable
					})
					if err != nil {
						return err
					}
				}
			}
		}
		// Check if there is any object blocking our path
		for _, o := range ctx.Data.Objects {
			if o.Name == object.Barrel && ctx.PathFinder.DistanceFromMe(o.Position) < 3 {
				err := step.InteractObject(o, func() bool {
					obj, found := ctx.Data.Objects.FindByID(o.ID)
					//additional click on barrel to avoid getting stuck
					x, y := ctx.PathFinder.GameCoordsToScreenCords(o.Position.X, o.Position.Y)
					ctx.HID.Click(game.LeftButton, x, y)
					return found && !obj.Selectable
				})
				if err != nil {
					return err
				}
			}
		}

		// Check for monsters close to player
		closestMonster := data.Monster{}
		closestMonsterDistance := 9999999
		targetedNormalEnemies := make([]data.Monster, 0)
		targetedElites := make([]data.Monster, 0)
		minDistance := 6
		minDistanceForElites := 20                                            // This will make the character to kill elites even if they are far away, ONLY during leveling
		stuck := ctx.PathFinder.DistanceFromMe(previousIterationPosition) < 5 // Detect if character was not able to move from last iteration

		for _, m := range ctx.Data.Monsters.Enemies() {
			// Skip if monster is already dead
			if m.Stats[stat.Life] <= 0 {
				continue
			}

			dist := ctx.PathFinder.DistanceFromMe(m.Position)
			appended := false
			if m.IsElite() && dist <= minDistanceForElites {
				targetedElites = append(targetedElites, m)
				appended = true
			}

			if dist <= minDistance {
				targetedNormalEnemies = append(targetedNormalEnemies, m)
				appended = true
			}

			if appended && dist < closestMonsterDistance {
				closestMonsterDistance = dist
				closestMonster = m
			}
		}

		if len(targetedNormalEnemies) > 5 || len(targetedElites) > 0 || (stuck && (len(targetedNormalEnemies) > 0 || len(targetedElites) > 0)) || (pather.IsNarrowMap(ctx.Data.PlayerUnit.Area) && (len(targetedNormalEnemies) > 0 || len(targetedElites) > 0)) {
			if stuck {
				ctx.Logger.Info("Character stuck and monsters detected, trying to kill monsters around")
			} else {
				ctx.Logger.Info(fmt.Sprintf("At least %d monsters detected close to the character, targeting closest one: %d", len(targetedNormalEnemies)+len(targetedElites), closestMonster.Name))
			}

			path, _, mPathFound := ctx.PathFinder.GetPath(closestMonster.Position)
			if mPathFound {
				doorIsBlocking := false
				for _, o := range ctx.Data.Objects {
					if o.IsDoor() && o.Selectable && path.Intersects(*ctx.Data, o.Position, 4) {
						ctx.Logger.Debug("Door is blocking the path to the monster, skipping attack sequence")
						doorIsBlocking = true
					}
				}

				if !doorIsBlocking {
					ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
						return closestMonster.UnitID, true
					}, nil)
				}
			}
		}

		// Continue moving
		WaitForAllMembersWhenLeveling()
		previousIterationPosition = ctx.Data.PlayerUnit.Position

		if lastMovement {
			return nil
		}

		// TODO: refactor this to use the same approach as ClearThroughPath
		if _, distance, _ := ctx.PathFinder.GetPathFrom(ctx.Data.PlayerUnit.Position, to); distance <= step.DistanceToFinishMoving {
			lastMovement = true
		}

		err := step.MoveTo(to)
		if err != nil {
			return err
		}
	}
}
