package pather

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Pather []astar.Pather

func (p Pather) Distance() int {
	return len(p)
}

// Intersects checks if the given position intersects with the path, padding parameter is used to increase the area
func (p Pather) Intersects(d game.Data, position data.Position, padding int) bool {
	position = data.Position{
		X: position.X - d.AreaOrigin.X,
		Y: position.Y - d.AreaOrigin.Y,
	}

	for _, path := range p {
		pT := path.(*Tile)
		xMatch := false
		yMatch := false
		for i := range padding {
			if pT.X == position.X+i || pT.X == position.X-i {
				xMatch = true
			}
			if pT.Y == position.Y+i || pT.Y == position.Y-i {
				yMatch = true
			}
		}

		if xMatch && yMatch {
			return true
		}
	}

	return false
}

func (p Pather) To() data.Position {
	if len(p) == 0 {
		return data.Position{}
	}

	return data.Position{X: p[len(p)-1].(*Tile).X, Y: p[len(p)-1].(*Tile).Y}
}
