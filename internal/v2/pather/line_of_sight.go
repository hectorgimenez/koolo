package pather

import (
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
)

func (pf *PathFinder) LineOfSight(origin data.Position, destination data.Position) bool {
	x0, y0 := origin.X-pf.data.AreaOrigin.X, origin.Y-pf.data.AreaOrigin.Y
	x1, y1 := destination.X-pf.data.AreaOrigin.X, destination.Y-pf.data.AreaOrigin.Y

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
		// Boundary check for x0, y0
		if x0 < 0 || y0 < 0 || x0 >= len(pf.data.CollisionGrid[0]) || y0 >= len(pf.data.CollisionGrid) {
			return false
		}

		// Check if the current position is not walkable
		if !pf.data.CollisionGrid[y0][x0] {
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
