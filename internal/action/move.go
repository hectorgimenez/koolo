package action

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/event"
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

//TODO something is off with Diablo action.MoveToArea(area.ChaosSanctuary) , the transition between movetoarea and clearthrough path makes
// bot telestomp until timeout before resuming with the run

func MoveToArea(dst area.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("MoveToArea")

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

	to, found := toFunc()
	if !found {
		return nil
	}

	// If we can teleport, just move straight to destination
	if ctx.Data.CanTeleport() {
		return step.MoveTo(to)
	}

	openedDoors := make(map[object.Name]data.Position)
	lastMovement := false

	for {
		ctx.PauseIfNotPriority()

		currentPos := ctx.Data.PlayerUnit.Position

		// Get current path segment
		var path []data.Position
		var found bool
		if ctx.CurrentGame.PathCache != nil &&
			ctx.CurrentGame.PathCache.DestPosition == to &&
			step.IsPathValid(currentPos, ctx.CurrentGame.PathCache) {
			path = ctx.CurrentGame.PathCache.Path
			found = true
		} else {

			path, _, found = ctx.PathFinder.GetPath(to)
			if found {
				ctx.CurrentGame.PathCache = &context.PathCache{
					Path:             path,
					DestPosition:     to,
					StartPosition:    currentPos,
					DistanceToFinish: step.DistanceToFinishMoving,
				}
			}
		}

		if !found {
			return fmt.Errorf("path could not be calculated")
		}

		//TODO if character detects monster on other side of the wall and we are past door detection
		// it will ignore doors and try to move directly to it. Either line of sight fail or constantly check for doors intersects

		// Walkable logic
		// Handle obstacles in current area
		if !ctx.Data.CanTeleport() {
			// Handle doors in path
			for _, o := range ctx.Data.Objects {
				if o.IsDoor() && o.Selectable && openedDoors[o.Name] != o.Position {
					// Check if door is near our path
					doorInPath := false
					doorPos := data.Position{
						X: o.Position.X - ctx.Data.AreaOrigin.X,
						Y: o.Position.Y - ctx.Data.AreaOrigin.Y,
					}

					for _, pathPos := range path {
						if pather.DistanceFromPoint(doorPos, pathPos) < 5 {
							doorInPath = true
							break
						}
					}

					if doorInPath && ctx.PathFinder.DistanceFromMe(o.Position) < 10 {
						ctx.Logger.Info("Door detected in path, attempting to open...")
						openedDoors[o.Name] = o.Position
						if err := step.InteractObject(o, func() bool {
							obj, found := ctx.Data.Objects.FindByID(o.ID)
							return found && !obj.Selectable
						}); err != nil {
							return err
						}

						// Clear area after opening door
						if err := ClearAreaAroundPosition(o.Position, 5, data.MonsterAnyFilter()); err != nil {
							ctx.Logger.Debug("Failed to clear area around door", err)
						}
						utils.Sleep(200)
						continue
					}
				}
			}
			// Clear monsters around current position
			if err := ClearAreaAroundPlayer(7, data.MonsterAnyFilter()); err != nil {
				return err
			}
		}
		// Continue moving

		// WaitForAllMembersWhenLeveling  is breaking walkable logic, lets not use it for now.

		if lastMovement {
			return nil
		}

		// Calculate next movement segment
		movementDistance := 7
		if movementDistance > len(path) {
			movementDistance = len(path)
		}

		dest := data.Position{
			X: path[movementDistance-1].X + ctx.Data.AreaOrigin.X,
			Y: path[movementDistance-1].Y + ctx.Data.AreaOrigin.Y,
		}

		if len(path)-movementDistance <= step.DistanceToFinishMoving {
			lastMovement = true
		}

		if err := step.MoveTo(dest); err != nil {
			return err
		}
	}
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
