package pather

import (
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
)

// parseWorld parses a textual representation of a World into a World map.
func parseWorld(collisionGrid [][]bool, d *game.Data) World {
	gridSizeX := len(collisionGrid[0])
	gridSizeY := len(collisionGrid)

	w := World{
		World: make([][]*Tile, gridSizeX),
	}
	for x := 0; x < gridSizeX; x++ {
		w.World[x] = make([]*Tile, gridSizeY)
	}

	for x, xValues := range collisionGrid {
		for y, walkable := range xValues {
			kind := KindBlocker

			// Hacky solution to avoid Arcane Sanctuary A* errors
			if d.PlayerUnit.Area == area.ArcaneSanctuary && d.CanTeleport() {
				kind = KindSoftBlocker
			}

			if walkable {
				// Add some padding around non-walkable areas, this prevents problems when cornering without teleport
				if !d.CanTeleport() && ((y > 1 && (!xValues[y-1] || !xValues[y-2])) || (y < len(xValues)-2 && (!xValues[y+1] || !xValues[y+2])) ||
					(x > 1 && (!collisionGrid[x-1][y] || !collisionGrid[x-2][y])) || (x < len(collisionGrid)-2 && (!collisionGrid[x+1][y] || !collisionGrid[x+2][y]))) {
					kind = KindSoftBlocker
				} else {
					kind = KindPlain
				}
			}

			w.SetTile(w.NewTile(kind, y, x))
		}
	}

	return w
}

func IsNarrowMap(a area.ID) bool {
	switch a {
	case area.MaggotLairLevel1, area.MaggotLairLevel2, area.MaggotLairLevel3, area.ArcaneSanctuary, area.ClawViperTempleLevel2, area.RiverOfFlame, area.ChaosSanctuary:
		return true
	}

	return false
}

func DistanceFromPoint(from data.Position, to data.Position) int {
	first := math.Pow(float64(to.X-from.X), 2)
	second := math.Pow(float64(to.Y-from.Y), 2)

	return int(math.Sqrt(first + second))
}

func IsWalkable(pos data.Position, areaOriginPos data.Position, collisionGrid [][]bool) bool {
	indexX := pos.X - areaOriginPos.X
	indexY := pos.Y - areaOriginPos.Y

	// When we are close to the level border, we need to check if monster is outside the collision grid
	if indexX < 0 || indexY < 0 || indexY >= len(collisionGrid) || indexX >= len(collisionGrid[indexY]) {
		return false
	}

	return collisionGrid[indexY][indexX]
}

// FindFirstWalkable finds the first walkable position from a given position and radius
func FindFirstWalkable(from data.Position, areaOriginPos data.Position, grid [][]bool, radius int) (int, int) {
	startX := from.X - areaOriginPos.X
	startY := from.Y - areaOriginPos.Y

	for r := radius; r >= 0; r-- {
		for dx := -r; dx <= r; dx++ {
			dy := int(math.Sqrt(float64(r*r - dx*dx)))
			positions := [][2]int{
				{startX + dx, startY + dy},
				{startX + dx, startY - dy},
				{startX - dx, startY + dy},
				{startX - dx, startY - dy},
			}
			for _, pos := range positions {
				newX, newY := pos[0]+areaOriginPos.X, pos[1]+areaOriginPos.Y
				if pos[0] >= 0 && pos[0] < len(grid) && pos[1] >= 0 && pos[1] < len(grid[0]) && IsWalkable(data.Position{newX, newY}, areaOriginPos, grid) {
					return newX, newY
				}
			}
		}
	}
	return -1, -1
}
