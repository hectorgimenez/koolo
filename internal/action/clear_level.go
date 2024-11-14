package action

import (
	"errors"
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ClearCurrentLevel(openChests bool, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.SetLastAction("ClearCurrentLevel")

	// First check if we're in town, if so, exit early
	if ctx.Data.PlayerUnit.Area.IsTown() {
		return fmt.Errorf("cannot clear level in town")
	}

	currentArea := ctx.Data.PlayerUnit.Area
	rooms := ctx.PathFinder.OptimizeRoomsTraverseOrder()
	for _, r := range rooms {
		// If we're no longer in the same area (e.g., took portal to town), stop clearing
		if ctx.Data.PlayerUnit.Area != currentArea {
			return nil
		}

		err := clearRoom(r, filter)
		if err != nil {
			ctx.Logger.Warn("Failed to clear room", "error", err)
		}

		if !openChests {
			continue
		}

		// Again check area before chest operations
		if ctx.Data.PlayerUnit.Area != currentArea {
			return nil
		}

		for _, o := range ctx.Data.Objects {
			if o.IsChest() && o.Selectable && r.IsInside(o.Position) {
				err = MoveToCoords(o.Position)
				if err != nil {
					ctx.Logger.Warn("Failed moving to chest", "error", err)
					continue
				}
				err = InteractObject(o, func() bool {
					chest, _ := ctx.Data.Objects.FindByID(o.ID)
					return !chest.Selectable
				})
				if err != nil {
					ctx.Logger.Warn("Failed interacting with chest", "error", err)
				}
				utils.Sleep(500) // Add small delay to allow the game to open the chest and drop the content
			}
		}
	}

	return nil
}
func clearRoom(room data.Room, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.SetLastAction("clearRoom")

	currentArea := ctx.Data.PlayerUnit.Area

	path, _, found := ctx.PathFinder.GetClosestWalkablePath(room.GetCenter())
	if !found {
		return errors.New("failed to find a path to the room center")
	}

	to := data.Position{
		X: path.To().X + ctx.Data.AreaOrigin.X,
		Y: path.To().Y + ctx.Data.AreaOrigin.Y,
	}

	// Check if we're still in the same area before moving
	if ctx.Data.PlayerUnit.Area != currentArea {
		return nil
	}

	err := MoveToCoords(to)
	if err != nil {
		return fmt.Errorf("failed moving to room center: %w", err)
	}

	for {
		// Exit if we've changed areas
		if ctx.Data.PlayerUnit.Area != currentArea {
			return nil
		}

		monsters := getMonstersInRoom(room, filter)
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

func getMonstersInRoom(room data.Room, filter data.MonsterFilter) []data.Monster {
	ctx := context.Get()
	ctx.SetLastAction("getMonstersInRoom")

	monstersInRoom := make([]data.Monster, 0)
	for _, m := range ctx.Data.Monsters.Enemies(filter) {
		if m.Stats[stat.Life] > 0 && room.IsInside(m.Position) || ctx.PathFinder.DistanceFromMe(m.Position) < 30 {
			monstersInRoom = append(monstersInRoom, m)
		}
	}

	return monstersInRoom
}
