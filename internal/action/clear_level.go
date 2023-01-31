package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/pather"
)

func (b Builder) ClearArea(openChests bool, filter game.MonsterFilter) *Factory {
	var clearedRooms []game.Room

	return NewFactory(func(data game.Data) Action {
		var currentRoom game.Room
		for _, r := range data.Rooms {
			if r.IsInside(data.PlayerUnit.Position) {
				currentRoom = r
				break
			}
		}

		monstersInRoom := make([]game.Monster, 0)
		for _, m := range data.Monsters.Enemies(filter) {
			if currentRoom.IsInside(m.Position) {
				monstersInRoom = append(monstersInRoom, m)
			}
		}

		if len(monstersInRoom) > 0 {
			return b.ch.KillMonsterSequence(func(data game.Data) (game.UnitID, bool) {
				return monstersInRoom[0].UnitID, true
			}, nil, step.Distance(5, 15))
		}

		if alreadyCleared(currentRoom, clearedRooms) {
			// Finished, all rooms are clear
			if len(clearedRooms) == len(data.Rooms) {
				b.logger.Debug("All the rooms for this level have been cleared, finishing run.")
				return nil
			}

			// Move to the closest room
			previousDistance := 999999
			closestRoom := game.Room{}
			for _, r := range data.Rooms {
				if alreadyCleared(r, clearedRooms) {
					continue
				}
				d := pather.DistanceFromMe(data, r.GetCenter())
				if d < previousDistance {
					previousDistance = d
					closestRoom = r
				}
			}

			return BuildStatic(func(data game.Data) []step.Step {
				b.logger.Debug("Room already cleared, moving to the next one.")
				_, _, found := pather.GetPath(data, closestRoom.GetCenter())
				// We don't need to be very precise, usually chests are not close to the map border tiles
				if !found && data.PlayerUnit.Area != area.LowerKurast {
					_, _, found = pather.GetClosestWalkablePath(data, closestRoom.GetCenter())
				}
				if !found {
					b.logger.Debug("Next room is not walkable, skipping it.")
					clearedRooms = append(clearedRooms, closestRoom)
					return []step.Step{}
				}

				return []step.Step{step.MoveTo(
					closestRoom.GetCenter().X,
					closestRoom.GetCenter().Y,
					true,
					step.ClosestWalkable(),
					step.StopAtDistance(5),
				)}
			})
		}

		clearedRooms = append(clearedRooms, currentRoom)

		// Open chests if are inside this room
		if openChests {
			for _, o := range data.Objects {
				if o.IsSuperChest() && o.Selectable && currentRoom.IsInside(o.Position) {
					chest := o
					return BuildStatic(func(data game.Data) []step.Step {
						return []step.Step{
							step.MoveTo(o.Position.X, o.Position.Y, true),
							step.InteractObject(o.Name, func(data game.Data) bool {
								for _, obj := range data.Objects {
									if obj.Name == chest.Name && obj.Position.X == chest.Position.X && obj.Position.Y == chest.Position.Y && obj.Selectable {
										return false
									}
								}

								return true
							}),
						}
					})
				}
			}
		}

		return b.ItemPickup(false, 60)
	})
}

func alreadyCleared(room game.Room, clearedRooms []game.Room) bool {
	for _, r := range clearedRooms {
		if r.X == room.X && r.Y == room.Y {
			return true
		}
	}

	return false
}
