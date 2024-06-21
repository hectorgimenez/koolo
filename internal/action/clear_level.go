package action

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b *Builder) ClearArea(openChests bool, filter data.MonsterFilter) *Chain {
	var clearedRooms []data.Room
	openedDoors := make(map[object.Name]data.Position)

	return NewChain(func(d game.Data) []Action {
		var currentRoom data.Room
		for _, r := range d.Rooms {
			if r.IsInside(d.PlayerUnit.Position) {
				currentRoom = r
				break
			}
		}

		// Check if there is a door blocking our path
		if !d.CanTeleport() {
			for _, o := range d.Objects {
				if o.IsDoor() && pather.DistanceFromMe(d, o.Position) < 10 && openedDoors[o.Name] != o.Position {
					if o.Selectable {
						b.Logger.Info("Door detected and teleport is not available, trying to open it...")
						return []Action{b.InteractObject(o.Name, func(d game.Data) bool {
							for _, obj := range d.Objects {
								if obj.Name == o.Name && obj.Position == o.Position && !obj.Selectable {
									openedDoors[o.Name] = o.Position
									return true
								}
							}
							return false
						})}
					}
				}
			}
		}

		monstersInRoom := make([]data.Monster, 0)
		for _, m := range d.Monsters.Enemies(filter) {
			if currentRoom.IsInside(m.Position) || pather.DistanceFromMe(d, m.Position) < 30 {
				monstersInRoom = append(monstersInRoom, m)
			}
		}

		if len(monstersInRoom) > 0 {
			targetMonster := monstersInRoom[0]
			// Check if there are monsters that can summon new monsters, and kill them first
			for _, m := range monstersInRoom {
				if m.IsMonsterRaiser() {
					targetMonster = m
				}
			}

			path, _, mPathFound := b.PathFinder.GetPath(d, targetMonster.Position)
			if mPathFound {
				doorIsBlocking := false
				if !d.CanTeleport() {
					for _, o := range d.Objects {
						if o.IsDoor() && o.Selectable && path.Intersects(d, o.Position, 4) {
							b.Logger.Debug("Door is blocking the path to the monster, skipping attack sequence")
							doorIsBlocking = true
						}
					}
				}

				if !doorIsBlocking {
					return []Action{b.ch.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
						m, found := d.Monsters.FindByID(targetMonster.UnitID)
						if found && m.Stats[stat.Life] > 0 {
							return targetMonster.UnitID, true
						}
						return 0, false
					}, nil)}
				} else {
					b.Logger.Debug("Door is blocking the path to the monster, skipping attack sequence")
				}
			}
		}

		if alreadyCleared(currentRoom, clearedRooms) {
			// Finished, all rooms are clear
			if len(clearedRooms) == len(d.Rooms) {
				b.Logger.Debug("All the rooms for this level have been cleared, finishing run.")
				return nil
			}

			// Move to the closest room
			previousDistance := 999999
			closestRoom := data.Room{}
			for _, r := range d.Rooms {
				if alreadyCleared(r, clearedRooms) {
					continue
				}
				d := pather.DistanceFromMe(d, r.GetCenter())
				if d < previousDistance {
					previousDistance = d
					closestRoom = r
				}
			}

			return []Action{NewStepChain(func(d game.Data) []step.Step {
				_, distance, found := b.PathFinder.GetPath(d, closestRoom.GetCenter())
				// We don't need to be very precise, usually chests are not close to the map border tiles
				if !found && d.PlayerUnit.Area != area.LowerKurast {
					_, distance, found = b.PathFinder.GetClosestWalkablePath(d, closestRoom.GetCenter())
				}
				if !found || distance <= 5 {
					b.Logger.Debug("Next room is not walkable, skipping it.")
					clearedRooms = append(clearedRooms, closestRoom)
					return []step.Step{}
				}

				return []step.Step{step.MoveTo(
					closestRoom.GetCenter(),
					step.WithTimeout(time.Second),
				)}
			})}
		}

		clearedRooms = append(clearedRooms, currentRoom)

		// Open chests if are inside this room
		if openChests {
			for _, o := range d.Objects {
				if o.IsChest() && o.Selectable && currentRoom.IsInside(o.Position) {
					chest := o
					doorIsBlocking := false
					if !d.CanTeleport() {
						path, _, chestPathFound := b.PathFinder.GetPath(d, chest.Position)
						for _, obj := range d.Objects {
							if chestPathFound && obj.IsDoor() && obj.Selectable && path.Intersects(d, obj.Position, 4) {
								doorIsBlocking = true
							}
						}
					}
					if !doorIsBlocking {
						return []Action{b.InteractObjectByID(chest.ID, func(d game.Data) bool {
							chest, _ = d.Objects.FindByID(chest.ID)
							return !chest.Selectable
						})}
					}
				}
			}
		}

		return []Action{b.ItemPickup(false, 60)}
	}, RepeatUntilNoSteps())
}

func alreadyCleared(room data.Room, clearedRooms []data.Room) bool {
	for _, r := range clearedRooms {
		if r.X == room.X && r.Y == room.Y {
			return true
		}
	}

	return false
}
