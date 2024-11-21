package pather

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather/astar"
)

type PathFinder struct {
	gr   *game.MemoryReader
	data *game.Data
	hid  *game.HID
	cfg  *config.CharacterCfg
}

func NewPathFinder(gr *game.MemoryReader, data *game.Data, hid *game.HID, cfg *config.CharacterCfg) *PathFinder {
	return &PathFinder{
		gr:   gr,
		data: data,
		hid:  hid,
		cfg:  cfg,
	}
}

type areaTransition struct {
	from, to area.ID
}

func (pf *PathFinder) isWallEntrance(from, to area.ID) bool {
	// Act 1
	if from == area.DarkWood && to == area.UndergroundPassageLevel1 {
		return true
	}
	if from == area.Cathedral && to == area.CatacombsLevel1 {
		return true
	}
	if from == area.JailLevel3 && to == area.InnerCloister {
		return true
	}

	// Act 2
	if from == area.StonyTombLevel1 && to == area.StonyTombLevel2 {
		return true
	}
	if from == area.HallsOfTheDeadLevel1 && to == area.HallsOfTheDeadLevel2 {
		return true
	}
	if from == area.ClawViperTempleLevel1 && to == area.ClawViperTempleLevel2 {
		return true
	}
	if from == area.CanyonOfTheMagi && (to >= area.TalRashasTomb1 && to <= area.TalRashasTomb7) {
		return true
	}
	// Handle Duriel's Lair from any Tal Rasha Tomb
	if (from >= area.TalRashasTomb1 && from <= area.TalRashasTomb7) && to == area.DurielsLair {
		return true
	}

	// Act 5
	if from == area.CrystallinePassage && to == area.ArreatPlateau {
		return true
	}
	if from == area.ArreatPlateau && to == area.GlacialTrail {
		return true
	}
	if from == area.GlacialTrail && to == area.FrozenTundra {
		return true
	}

	return false
}

func (pf *PathFinder) findNearbyWalkablePosition(target data.Position) (data.Position, bool) {

	for radius := 1; radius <= 3; radius++ {
		for x := -radius; x <= radius; x++ {
			for y := -radius; y <= radius; y++ {
				if x == 0 && y == 0 {
					continue
				}
				pos := data.Position{X: target.X + x, Y: target.Y + y}
				if pf.data.AreaData.IsWalkable(pos) {
					return pos, true
				}
			}
		}
	}

	return data.Position{}, false
}

func (pf *PathFinder) GetPath(to data.Position) (Path, int, bool) {
	// Validate target position is either in current area or in an adjacent open area
	// IsInside will prevent attempts to path to coordinates from a different area
	if !pf.data.AreaData.IsInside(to) {
		// Check if it's in any adjacent level before failing
		validPosition := false
		for _, level := range pf.data.AdjacentLevels {
			if !level.IsEntrance && pf.data.Areas[level.Area].IsInside(to) {
				validPosition = true
				break
			}
		}
		if !validPosition {
			return nil, 0, false
		}
	}

	// Special case for duriels lair entrance
	for _, obj := range pf.data.Objects {
		if obj.Name == object.DurielsLairPortal {
			if walkable, found := pf.findNearbyWalkablePosition(to); found {
				return pf.GetPathFrom(pf.data.PlayerUnit.Position, walkable)
			}
		}
	}

	// Then check for wall entrances
	for _, level := range pf.data.AdjacentLevels {
		if level.IsEntrance && level.Position == to && pf.isWallEntrance(pf.data.PlayerUnit.Area, level.Area) {
			if walkable, found := pf.findNearbyWalkablePosition(to); found {
				return pf.GetPathFrom(pf.data.PlayerUnit.Position, walkable)
			}
		}
	}

	// Normal pathing for non-entrance destinations
	return pf.GetPathFrom(pf.data.PlayerUnit.Position, to)
}

func (pf *PathFinder) GetPathFrom(from, to data.Position) (Path, int, bool) {
	a := pf.data.AreaData

	// We don't want to modify the original grid
	grid := a.Grid.Copy()

	// Lut Gholein map is a bit bugged, we should close this fake path to avoid pathing issues
	if a.Area == area.LutGholein {
		a.CollisionGrid[13][210] = game.CollisionTypeNonWalkable
	}

	if !a.IsInside(to) {
		expandedGrid, err := pf.mergeGrids(to)
		if err != nil {
			return nil, 0, false
		}
		grid = expandedGrid
	}

	from = grid.RelativePosition(from)
	to = grid.RelativePosition(to)

	// Add objects to the collision grid as obstacles
	for _, o := range pf.data.AreaData.Objects {
		if !grid.IsWalkable(o.Position) {
			continue
		}
		relativePos := grid.RelativePosition(o.Position)
		grid.CollisionGrid[relativePos.Y][relativePos.X] = game.CollisionTypeObject
		for i := -2; i <= 2; i++ {
			for j := -2; j <= 2; j++ {
				if i == 0 && j == 0 {
					continue
				}
				if relativePos.Y+i < 0 || relativePos.Y+i >= len(grid.CollisionGrid) || relativePos.X+j < 0 || relativePos.X+j >= len(grid.CollisionGrid[relativePos.Y]) {
					continue
				}
				if grid.CollisionGrid[relativePos.Y+i][relativePos.X+j] == game.CollisionTypeWalkable {
					grid.CollisionGrid[relativePos.Y+i][relativePos.X+j] = game.CollisionTypeLowPriority

				}
			}
		}
	}

	// Add monsters to the collision grid as obstacles
	for _, m := range pf.data.Monsters {
		if !grid.IsWalkable(m.Position) {
			continue
		}
		relativePos := grid.RelativePosition(m.Position)
		grid.CollisionGrid[relativePos.Y][relativePos.X] = game.CollisionTypeMonster
	}

	path, distance, found := astar.CalculatePath(grid, from, to)

	if config.Koolo.Debug.RenderMap {
		pf.renderMap(grid, from, to, path)
	}

	return path, distance, found
}

func (pf *PathFinder) mergeGrids(to data.Position) (*game.Grid, error) {
	for _, a := range pf.data.AreaData.AdjacentLevels {
		destination := pf.data.Areas[a.Area]
		if destination.IsInside(to) {
			origin := pf.data.AreaData

			endX1 := origin.OffsetX + len(origin.Grid.CollisionGrid[0])
			endY1 := origin.OffsetY + len(origin.Grid.CollisionGrid)
			endX2 := destination.OffsetX + len(destination.Grid.CollisionGrid[0])
			endY2 := destination.OffsetY + len(destination.Grid.CollisionGrid)

			minX := min(origin.OffsetX, destination.OffsetX)
			minY := min(origin.OffsetY, destination.OffsetY)
			maxX := max(endX1, endX2)
			maxY := max(endY1, endY2)

			width := maxX - minX
			height := maxY - minY

			resultGrid := make([][]game.CollisionType, height)
			for i := range resultGrid {
				resultGrid[i] = make([]game.CollisionType, width)
			}

			// Let's copy both grids into the result grid
			copyGrid(resultGrid, origin.CollisionGrid, origin.OffsetX-minX, origin.OffsetY-minY)
			copyGrid(resultGrid, destination.CollisionGrid, destination.OffsetX-minX, destination.OffsetY-minY)

			grid := game.NewGrid(resultGrid, minX, minY)

			return grid, nil
		}
	}

	return nil, fmt.Errorf("destination grid not found")
}

func copyGrid(dest [][]game.CollisionType, src [][]game.CollisionType, offsetX, offsetY int) {
	for y := 0; y < len(src); y++ {
		for x := 0; x < len(src[0]); x++ {
			dest[offsetY+y][offsetX+x] = src[y][x]
		}
	}
}

func (pf *PathFinder) GetClosestWalkablePath(dest data.Position, maxRange int) (Path, int, bool) {
	return pf.GetClosestWalkablePathFrom(pf.data.PlayerUnit.Position, dest, maxRange)
}

func (pf *PathFinder) GetClosestWalkablePathFrom(from, dest data.Position, maxRange int) (Path, int, bool) {
	a := pf.data.AreaData
	if a.IsWalkable(dest) || !a.IsInside(dest) {
		// Directly use GetPathFrom instead of GetPath to avoid recursion
		return pf.GetPathFrom(from, dest)
	}

	// If no maxRange specified, use default of 20
	if maxRange <= 0 {
		maxRange = 20
	}

	step := 4
	dst := 1

	for dst < maxRange {
		for i := -dst; i < dst; i += 1 {
			for j := -dst; j < dst; j += 1 {
				if math.Abs(float64(i)) >= math.Abs(float64(dst)) || math.Abs(float64(j)) >= math.Abs(float64(dst)) {
					cgY := dest.Y - pf.data.AreaOrigin.Y + j
					cgX := dest.X - pf.data.AreaOrigin.X + i
					if cgX > 0 && cgY > 0 && a.Height > cgY && a.Width > cgX && a.CollisionGrid[cgY][cgX] == game.CollisionTypeWalkable {
						return pf.GetPathFrom(from, data.Position{
							X: dest.X + i,
							Y: dest.Y + j,
						})
					}
				}
			}
		}
		dst += step
	}

	return nil, 0, false
}
