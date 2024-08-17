package pather

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (pf *PathFinder) GetWalkableRooms(d game.Data) []data.Position {
	roomPositions := make([]data.Position, 0)

	for _, room := range d.Rooms {
		found := false
		for y := range room.Height - 1 {
			for x := range room.Width - 1 {
				if d.CollisionGrid[room.Y-d.AreaOrigin.Y+y][room.X-d.AreaOrigin.X+x] {
					pos := data.Position{
						X: room.Position.X + x,
						Y: room.Position.Y + y,
					}
					if _, _, found = pf.GetPath(pos); found {
						roomPositions = append(roomPositions, pos)
						break
					}
				}
			}
			if found {
				break
			}
		}
	}

	return roomPositions
}
