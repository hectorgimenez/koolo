package action

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b *Builder) ClearArea(openChests bool, filter data.MonsterFilter) *Chain {
	var clearedRooms []data.Room
	openedDoors := make(map[object.Name]data.Position)

	return NewChain(func(d data.Data) []Action {
		var currentRoom data.Room
		for _, r := range d.Rooms {
			if r.IsInside(d.PlayerUnit.Position) {
				currentRoom = r
				break
			}
		}

		// Let's go pickup more pots if we have less than 2 (only during leveling)
		_, isLevelingChar := b.ch.(LevelingCharacter)
		if isLevelingChar {
			_, healingPotsFound := d.Items.Belt.GetFirstPotion(data.HealingPotion)
			_, manaPotsFound := d.Items.Belt.GetFirstPotion(data.ManaPotion)
			if (!healingPotsFound || !manaPotsFound) && d.PlayerUnit.TotalGold() > 1000 {
				return b.InRunReturnTownRoutine()
			}
		}

		// Check if there is a door blocking our path
		if !helper.CanTeleport(d) {
			for _, o := range d.Objects {
				if o.IsDoor() && pather.DistanceFromMe(d, o.Position) < 10 && openedDoors[o.Name] != o.Position {
					if o.Selectable {
						b.logger.Info("Door detected and teleport is not available, trying to open it...")
						return []Action{b.InteractObject(o.Name, func(d data.Data) bool {
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

			path, _, mPathFound := pather.GetPath(d, targetMonster.Position)
			if mPathFound {
				doorIsBlocking := false
				if !helper.CanTeleport(d) {
					for _, o := range d.Objects {
						if o.IsDoor() && o.Selectable && path.Intersects(d, o.Position, 4) {
							b.logger.Debug("Door is blocking the path to the monster, skipping attack sequence")
							doorIsBlocking = true
						}
					}
				}

				if !doorIsBlocking {
					return []Action{b.ch.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
						return targetMonster.UnitID, true
					}, nil)}
				} else {
					b.logger.Debug("Door is blocking the path to the monster, skipping attack sequence")
				}
			}
		}

		if alreadyCleared(currentRoom, clearedRooms) {
			// Finished, all rooms are clear
			if len(clearedRooms) == len(d.Rooms) {
				b.logger.Debug("All the rooms for this level have been cleared, finishing run.")
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

			return []Action{NewStepChain(func(d data.Data) []step.Step {
				_, distance, found := pather.GetPath(d, closestRoom.GetCenter())
				// We don't need to be very precise, usually chests are not close to the map border tiles
				if !found && d.PlayerUnit.Area != area.LowerKurast {
					_, distance, found = pather.GetClosestWalkablePath(d, closestRoom.GetCenter())
				}
				if !found || distance <= 5 {
					b.logger.Debug("Next room is not walkable, skipping it.")
					clearedRooms = append(clearedRooms, closestRoom)
					return []step.Step{}
				}

				return []step.Step{step.MoveTo(
					closestRoom.GetCenter(),
					step.ClosestWalkable(),
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
					return []Action{b.InteractObject(chest.Name, func(d data.Data) bool {
						for _, obj := range d.Objects {
							if obj.Name == chest.Name && obj.Position.X == chest.Position.X && obj.Position.Y == chest.Position.Y && obj.Selectable {
								return false
							}
						}

						return true
					})}
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
