package pather

import "github.com/beefsack/go-astar"

var allowedMovement = [][]int{
	{-1, 0},
	{1, 0},
	{0, -1},
	{0, 1},
}

// Kind* constants refer to tile kinds for input and output.
const (
	// KindPlain (.) is a plain tile with a movement cost of 1.
	KindPlain uint8 = iota
	// KindSoftBlocker (S) is a tile to fake some blocking areas with increased cost of moving, like Arcane Sanctuary platforms
	KindSoftBlocker
	// KindBlocker (X) is a tile which blocks movement.
	KindBlocker
)

// KindCosts map tile kinds to movement costs.
var KindCosts = map[uint8]float64{
	KindPlain:       1.0,
	KindSoftBlocker: 1000,
}

// A Tile is a tile in a grid which implements Pather.
type Tile struct {
	Walkable bool
	Cost     float64
	X, Y     int
	Kind     uint8
	W        *World
}

// PathNeighbors returns the neighbors of the tile, excluding blockers and
// tiles off the edge of the board.
func (t *Tile) PathNeighbors() []astar.Pather {
	if t == nil {
		return []astar.Pather{}
	}

	neighbors := make([]astar.Pather, 0, 4)
	for _, offset := range allowedMovement {
		if n := t.W.Tile(t.X+offset[0], t.Y+offset[1]); n != nil && n.Walkable {
			neighbors = append(neighbors, n)
		}
	}

	return neighbors
}

// PathNeighborCost returns the movement cost of the directly neighboring tile.
func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	return to.(*Tile).Cost
}

// PathEstimatedCost uses Manhattan distance to estimate orthogonal distance
// between non-adjacent nodes.
func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	toT := to.(*Tile)
	absX := toT.X - t.X
	if absX < 0 {
		absX = -absX
	}
	absY := toT.Y - t.Y
	if absY < 0 {
		absY = -absY
	}
	return float64(absX + absY)
}
