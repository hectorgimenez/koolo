package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/context"
)

func ClearCurrentLevel(openChests bool, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "ClearCurrentLevel"

	for _, r := range ctx.PathFinder.OptimizeRoomsTraverseOrder() {
		err := clearRoom(r, filter)
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
			}
		}
	}

	return nil
}

func clearRoom(room data.Room, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.ContextDebug.LastAction = "clearRoom"

	path, _, found := ctx.PathFinder.GetClosestWalkablePath(room.GetCenter())
	if !found {
		return errors.New("failed to find a path to the room center")
	}

	err := MoveToCoords(path.To())
	if err != nil {
		ctx.Logger.Warn("Failed moving to room center: %v", err)
	}

	for monsters := getMonstersInRoom(room, filter); len(monsters) > 0; {
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

	return nil
}

func getMonstersInRoom(room data.Room, filter data.MonsterFilter) []data.Monster {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "getMonstersInRoom"

	monstersInRoom := make([]data.Monster, 0)
	for _, m := range ctx.Data.Monsters.Enemies(filter) {
		if room.IsInside(m.Position) || ctx.PathFinder.DistanceFromMe(m.Position) < 30 {
			monstersInRoom = append(monstersInRoom, m)
		}
	}

	return monstersInRoom
}