package pather

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Path []data.Position

func (p Path) To() data.Position {
	return data.Position{
		X: p[len(p)-1].X,
		Y: p[len(p)-1].Y,
	}
}

func (p Path) From() data.Position {
	return data.Position{
		X: p[0].X,
		Y: p[0].Y,
	}
}

// Intersects checks if the given position intersects with the path, padding parameter is used to increase the area
func (p Path) Intersects(d game.Data, position data.Position, padding int) bool {
	position = data.Position{
		X: position.X - d.AreaOrigin.X,
		Y: position.Y - d.AreaOrigin.Y,
	}

	for _, point := range p {
		xMatch := false
		yMatch := false
		for i := range padding {
			if point.X == position.X+i || point.X == position.X-i {
				xMatch = true
			}
			if point.Y == position.Y+i || point.Y == position.Y-i {
				yMatch = true
			}
		}

		if xMatch && yMatch {
			return true
		}
	}

	return false
}
