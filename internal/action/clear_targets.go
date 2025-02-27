package action

import (
	"errors"
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ClearCurrentLevelTargets(openChests bool, ID npc.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("ClearCurrentLevel")

	rooms := ctx.PathFinder.OptimizeRoomsTraverseOrder()
	for _, r := range rooms {
		err := clearRoomID(r, ID)
		if err != nil {
			ctx.Logger.Warn("Failed to clear room: %v", err)
		}

		if !openChests {
			continue
		}

		for _, o := range ctx.Data.Objects {
			if o.IsChest() && o.Selectable && r.IsInside(o.Position) {
				err = MoveToCoords(o.Position)
				if err != nil {
					ctx.Logger.Warn("Failed moving to chest: %v", err)
					continue
				}
				err = InteractObject(o, func() bool {
					chest, _ := ctx.Data.Objects.FindByID(o.ID)
					return !chest.Selectable
				})
				if err != nil {
					ctx.Logger.Warn("Failed interacting with chest: %v", err)
				}
				utils.Sleep(500) // Add small delay to allow the game to open the chest and drop the content
			}
		}
	}

	return nil
}

func clearRoomID(room data.Room, ID npc.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("clearRoom")

	path, _, found := ctx.PathFinder.GetClosestWalkablePath(room.GetCenter())
	if !found {
		return errors.New("failed to find a path to the room center")
	}

	to := data.Position{
		X: path.To().X + ctx.Data.AreaOrigin.X,
		Y: path.To().Y + ctx.Data.AreaOrigin.Y,
	}
	err := MoveToCoords(to)
	if err != nil {
		return fmt.Errorf("failed moving to room center: %w", err)
	}

	for {
		monsters := getMonstersIDsInRoom(room, ID)
		if len(monsters) == 0 {
			return nil
		}

		// Check if there are monsters that can summon new monsters, and kill them first
		targetMonster := monsters[0]
		for _, m := range monsters {
			if m.IsMonsterRaiser() {
				targetMonster = m
			}
		}

		path, _, mPathFound := ctx.PathFinder.GetPath(targetMonster.Position)
		if mPathFound {
			if !ctx.Data.CanTeleport() {
				for _, o := range ctx.Data.Objects {
					if o.IsDoor() && o.Selectable && path.Intersects(*ctx.Data, o.Position, 4) {
						ctx.Logger.Debug("Door is blocking the path to the monster, moving closer")
						MoveToCoords(targetMonster.Position)
					}
				}
			}

			ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				m, found := d.Monsters.FindByID(targetMonster.UnitID)
				if found && m.Stats[stat.Life] > 0 {
					return targetMonster.UnitID, true
				}
				return 0, false
			}, nil)
		}
	}
}

func getMonstersIDsInRoom(room data.Room, ID npc.ID) []data.Monster {
	ctx := context.Get()
	ctx.SetLastAction("getMonstersIDsInRoom")

	monstersInRoom := make([]data.Monster, 0)
	for _, m := range ctx.Data.Monsters.Enemies(data.MonsterAnyFilter()) {
		if m.Stats[stat.Life] > 0 && m.Name == ID && room.IsInside(m.Position) || ctx.PathFinder.DistanceFromMe(m.Position) < 30 && m.Name == ID {
			monstersInRoom = append(monstersInRoom, m)
		}
	}

	return monstersInRoom
}

func ClearAreaTargetsAroundPosition(pos data.Position, radius int, ID npc.ID) error {
	ctx := context.Get()
	ctx.SetLastAction("ClearAreaTargetsAroundPosition")

	for {
		ctx.PauseIfNotPriority()

		// Check for closest monster within radius - monsters are already sorted by distance
		err := ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies(data.MonsterAnyFilter()) {
				dist := pather.DistanceFromPoint(pos, m.Position)
				if ctx.Data.AreaData.IsWalkable(m.Position) && dist <= radius && m.Name == ID {
					return m.UnitID, true
				}
			}
			return 0, false
		}, nil)

		if err != nil {
			return err
		}

		// If no monsters found within radius, we're done
		found := false
		for _, m := range ctx.Data.Monsters.Enemies(data.MonsterAnyFilter()) {
			if pather.DistanceFromPoint(pos, m.Position) <= radius && m.Name == ID {
				MoveToCoords(m.Position)
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
}

func ClearTargetsThroughPath(pos data.Position, radius int, ID npc.ID) error {
	ctx := context.Get()

	lastMovement := false
	for {
		ctx.PauseIfNotPriority()

		ClearAreaTargetsAroundPosition(ctx.Data.PlayerUnit.Position, radius, ID)

		if lastMovement {
			return nil
		}

		path, _, found := ctx.PathFinder.GetPath(pos)
		if !found {
			return fmt.Errorf("path could not be calculated")
		}

		movementDistance := radius
		if radius > len(path) {
			movementDistance = len(path)
		}

		dest := data.Position{
			X: path[movementDistance-1].X + ctx.Data.AreaData.OffsetX,
			Y: path[movementDistance-1].Y + ctx.Data.AreaData.OffsetY,
		}

		// Let's handle the last movement logic to MoveTo function, we will trust the pathfinder because
		// it can finish within a bigger distance than we expect (because blockers), so we will just check how far
		// we should be after the latest movement in a theoretical way
		if len(path)-movementDistance <= step.DistanceToFinishMoving {
			lastMovement = true
		}
		// Increasing DistanceToFinishMoving prevent not being to able to finish movement if our destination is center of a large object like Seal in diablo run.
		// is used only for pathing, attack.go will use default DistanceToFinishMoving
		err := step.MoveTo(dest, step.WithDistanceToFinish(7))
		if err != nil {
			return err
		}
	}
}
