package action

import (
	"fmt"
	"sort"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ClearCurrentLevel(openChests bool, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.SetLastAction("ClearCurrentLevel")

	rooms := ctx.PathFinder.OptimizeRoomsTraverseOrder()
	for _, r := range rooms {
		err := clearRoom(r, filter)
		if err != nil && err.Error() != "" {
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

func clearRoom(room data.Room, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.SetLastAction("clearRoom")

	if err := moveToRoomPosition(room); err != nil {
		return err
	}

	return clearRoomMonsters(room, filter)
}

func moveToRoomPosition(room data.Room) error {
	ctx := context.Get()

	center := room.GetCenter()

	// Try center position first
	if ctx.Data.AreaData.IsWalkable(center) {
		if err := MoveToCoords(center); err == nil {
			return nil
		}
	}

	// For room clearing, use larger radius bounded by room size
	maxRadius := min(room.Width/2, room.Height/2)
	if walkablePoint, found := ctx.PathFinder.FindNearbyWalkablePosition(center, maxRadius); found {
		if err := MoveToCoords(walkablePoint); err == nil {
			return nil
		}
	}

	return fmt.Errorf("") // No walkable position found in room but don't log it.
}

func clearRoomMonsters(room data.Room, filter data.MonsterFilter) error {
	ctx := context.Get()

	for {
		monsters := getMonstersInRoom(room, filter)
		if len(monsters) == 0 {
			return nil
		}

		sort.Slice(monsters, func(i, j int) bool {
			if monsters[i].IsMonsterRaiser() != monsters[j].IsMonsterRaiser() {
				return monsters[i].IsMonsterRaiser()
			}
			distI := ctx.PathFinder.DistanceFromMe(monsters[i].Position)
			distJ := ctx.PathFinder.DistanceFromMe(monsters[j].Position)
			return distI < distJ
		})

		for _, monster := range monsters {
			if !ctx.Data.AreaData.IsInside(monster.Position) {
				continue
			}

			// Check for door obstacles for non-teleporting characters
			if !ctx.Data.CanTeleport() {
				path, _, found := ctx.PathFinder.GetPath(monster.Position)
				if found {
					for _, o := range ctx.Data.Objects {
						if o.IsDoor() && o.Selectable && path.Intersects(*ctx.Data, o.Position, 4) {
							ctx.Logger.Debug("Door is blocking the path to the monster, moving closer")
							if err := MoveToCoords(monster.Position); err != nil {
								continue
							}
						}
					}
				}
			}

			if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, monster.Position) {
				// Let MoveToCoords handle the pathing (it uses GetPath)
				if err := MoveToCoords(monster.Position); err != nil {
					continue
				}
			}

			ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				m, found := d.Monsters.FindByID(monster.UnitID)
				if found && m.Stats[stat.Life] > 0 {
					return monster.UnitID, true
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
		// Skip dead monsters or those outside current area
		if m.Stats[stat.Life] <= 0 || !ctx.Data.AreaData.IsInside(m.Position) {
			continue
		}

		inRoom := room.IsInside(m.Position)
		nearPlayer := ctx.PathFinder.DistanceFromMe(m.Position) < 30

		if !inRoom && !nearPlayer {
			roomCenter := room.GetCenter()
			distToRoom := pather.DistanceFromPoint(roomCenter, m.Position)
			if distToRoom < room.Width/2+15 || distToRoom < room.Height/2+15 {
				monstersInRoom = append(monstersInRoom, m)
				continue
			}
		}

		if inRoom || nearPlayer {
			monstersInRoom = append(monstersInRoom, m)
		}
	}

	return monstersInRoom
}
