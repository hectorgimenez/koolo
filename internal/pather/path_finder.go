package pather

import (
	"fmt"
	"math"
	"sync"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather/astar"
)

type PathFinder struct {
	gr           *game.MemoryReader
	data         *game.Data
	hid          *game.HID
	cfg          *config.CharacterCfg
	gridLock     sync.Mutex    // Protects grid state from concurrent access
	lastGrid     *game.Grid    // Cache last processed grid
	lastGridArea area.ID       // Track area for grid cache validation
}

func NewPathFinder(gr *game.MemoryReader, data *game.Data, hid *game.HID, cfg *config.CharacterCfg) *PathFinder {
	return &PathFinder{
		gr:   gr,
		data: data,
		hid:  hid,
		cfg:  cfg,
	}
}

func (pf *PathFinder) GetPath(to data.Position) (Path, int, bool) {
	currentPos := pf.data.PlayerUnit.Position
	currentArea := pf.data.PlayerUnit.Area

	// Check cache first for existing paths
	if path, found := getCachedPath(currentPos, to, currentArea); found {
		return path, len(path), true
	}

	// Calculate new path and cache it if found
	path, distance, found := pf.GetPathFrom(currentPos, to)
	if found {
		cachePath(currentPos, to, currentArea, path)
	}
	return path, distance, found
}

func (pf *PathFinder) GetPathFrom(from, to data.Position) (Path, int, bool) {
	a := pf.data.AreaData

	pf.gridLock.Lock()
	defer pf.gridLock.Unlock()

	// Use cached grid if available and still valid
	var grid *game.Grid
	if pf.lastGrid != nil && pf.lastGridArea == a.Area {
		grid = pf.lastGrid
	} else {
		grid = a.Grid.Copy()
		pf.preprocessGrid(grid)
		pf.lastGrid = grid
		pf.lastGridArea = a.Area
	}

	// Handle cross-area pathing by merging grids
	if !a.IsInside(to) {
		expandedGrid, err := pf.mergeGrids(to)
		if err != nil {
			return nil, 0, false
		}
		pf.preprocessGrid(expandedGrid)
		grid = expandedGrid
		pf.lastGrid = grid
		pf.lastGridArea = a.Area
	}

	from = grid.RelativePosition(from)
	to = grid.RelativePosition(to)

	// Validate positions are within grid bounds
	if from.X < 0 || from.X >= grid.Width || from.Y < 0 || from.Y >= grid.Height ||
		to.X < 0 || to.X >= grid.Width || to.Y < 0 || to.Y >= grid.Height {
		return nil, 0, false
	}

	path, _, found := astar.CalculatePath(grid, a.Area, from, to)

	if config.Koolo.Debug.RenderMap {
		pf.renderMap(grid, from, to, path)
	}

	return path, len(path), found
}

func (pf *PathFinder) preprocessGrid(grid *game.Grid) {
    a := pf.data.AreaData
	// Special handling for Arcane Sanctuary (allow teleport pathing over voids)
    if a.Area == area.ArcaneSanctuary && pf.data.CanTeleport() {
		// Make all non-walkable tiles into low priority tiles for teleport pathing
        for y := 0; y < len(grid.CollisionGrid); y++ {
            for x := 0; x < len(grid.CollisionGrid[y]); x++ {
                if grid.CollisionGrid[y][x] == game.CollisionTypeNonWalkable {
                    grid.CollisionGrid[y][x] = game.CollisionTypeLowPriority
                }
            }
        }
    }

	// Fix for broken path in Lut Gholein
    if a.Area == area.LutGholein {
        grid.CollisionGrid[13][210] = game.CollisionTypeNonWalkable
    }

	// Add objects to the collision grid as obstacles
    for _, o := range pf.data.AreaData.Objects {
        // Enhanced Hidden Stash handling with 5x5 collision blocking
        if string(o.Name) == "hidden stash" {
            relativePos := grid.RelativePosition(o.Position)
            // Block 5x5 area around stashes
            for dy := -2; dy <= 2; dy++ {
                for dx := -2; dx <= 2; dx++ {
                    y := relativePos.Y + dy
                    x := relativePos.X + dx
                    if y >= 0 && y < len(grid.CollisionGrid) && 
                       x >= 0 && x < len(grid.CollisionGrid[y]) {
                        grid.CollisionGrid[y][x] = game.CollisionTypeNonWalkable
                    }
                }
            }
            continue
        }

        // Existing object handling
        if !grid.IsWalkable(o.Position) {
            continue
        }
        relativePos := grid.RelativePosition(o.Position)
		// Mark object position and create low priority area around it
        grid.CollisionGrid[relativePos.Y][relativePos.X] = game.CollisionTypeObject
        for i := -2; i <= 2; i++ {
            for j := -2; j <= 2; j++ {
                if i == 0 && j == 0 || relativePos.Y+i < 0 || 
                   relativePos.Y+i >= len(grid.CollisionGrid) || 
                   relativePos.X+j < 0 || relativePos.X+j >= len(grid.CollisionGrid[relativePos.Y]) {
                    continue
                }
                if grid.CollisionGrid[relativePos.Y+i][relativePos.X+j] == game.CollisionTypeWalkable {
                    grid.CollisionGrid[relativePos.Y+i][relativePos.X+j] = game.CollisionTypeLowPriority
                }
            }
        }
    }

    // Add monsters to the collision grid as high-cost obstacles
    for _, m := range pf.data.Monsters {
        if !grid.IsWalkable(m.Position) {
            continue
        }
        relativePos := grid.RelativePosition(m.Position)
        grid.CollisionGrid[relativePos.Y][relativePos.X] = game.CollisionTypeMonster
    }
}

// Combine adjacent level grids for cross-area pathfinding
func (pf *PathFinder) mergeGrids(to data.Position) (*game.Grid, error) {
	for _, a := range pf.data.AreaData.AdjacentLevels {
		destination := pf.data.Areas[a.Area]
		if destination.IsInside(to) {
			origin := pf.data.AreaData

			// Calculate merged grid dimensions
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

			// Copy both grids into the merged result grid
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

func (pf *PathFinder) GetClosestWalkablePath(dest data.Position) (Path, int, bool) {
	return pf.GetClosestWalkablePathFrom(pf.data.PlayerUnit.Position, dest)
}

// Find nearest accessible position when direct path is blocked
func (pf *PathFinder) GetClosestWalkablePathFrom(from, dest data.Position) (Path, int, bool) {
	a := pf.data.AreaData
	// First try direct path if destination is walkable or outside known area
	if a.IsWalkable(dest) || !a.IsInside(dest) {
		path, distance, found := pf.GetPath(dest)
		if found {
			return path, distance, found
		}
	}

	// Search in expanding squares around target position
	maxRange := 20
	step := 4
	dst := 1

	for dst < maxRange {
		for i := -dst; i < dst; i += 1 {
			for j := -dst; j < dst; j += 1 {
				// Check perimeter of current search radius
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
