package pather

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/d2go/pkg/data"
)

type Pather struct {
	AstarPather
	Destination data.Position
}

type AstarPather []astar.Pather

func (p *Pather) Distance() int {
	return len(p.AstarPather)
}

// Intersects checks if the given position intersects with the path, padding parameter is used to increase the area
func (p *Pather) Intersects(d data.Data, position data.Position, padding int) bool {
	position = data.Position{
		X: position.X - d.AreaOrigin.X,
		Y: position.Y - d.AreaOrigin.Y,
	}

	for _, path := range p.AstarPather {
		pT := path.(*Tile)
		xMatch := false
		yMatch := false
		for i := 0; i < padding; i++ {
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
