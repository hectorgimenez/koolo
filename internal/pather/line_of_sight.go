package pather

import (
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

func LineOfSight(d game.Data, origin data.Position, destination data.Position) bool {
	x0, y0 := origin.X-d.AreaOrigin.X, origin.Y-d.AreaOrigin.Y
	x1, y1 := destination.X-d.AreaOrigin.X, destination.Y-d.AreaOrigin.Y

	dx := int(math.Abs(float64(x1 - x0)))
	dy := int(math.Abs(float64(y1 - y0)))
	sx := -1
	sy := -1

	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}

	err := dx - dy

	for {
		// Check if the current position is not walkable
		if !d.CollisionGrid[y0][x0] {
			return false
		}

		// Check if we have reached the end point
		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}

	return true
}
