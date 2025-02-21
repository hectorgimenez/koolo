package pather

import (
	"fmt"
	"log"
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
	gridLock     sync.Mutex // Protects grid state from concurrent access
	lastGrid     *game.Grid // Cache last processed grid
	lastGridArea area.ID    // Track area for grid cache validation
	lastTeleport bool       // Track last teleport state for cache validation
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
	teleportEnabled := pf.data.CanTeleport()

	// Check cache with teleport status
	if path, found := getCachedPath(currentPos, to, currentArea, teleportEnabled); found {
		log.Printf("DEBUG: Using cached path from %v to %v", currentPos, to)
		return path, len(path), true
	}

	// Calculate and cache new path
	path, distance, found := pf.GetPathFrom(currentPos, to)
	if found {
		log.Printf("DEBUG: Caching new path from %v to %v (length: %d)", currentPos, to, len(path))
		cachePath(currentPos, to, currentArea, teleportEnabled, path)
	}
	return path, distance, found
}

func (pf *PathFinder) GetPathFrom(from, to data.Position) (Path, int, bool) {
	a := pf.data.AreaData

	teleportEnabled := pf.data.CanTeleport()
	pf.gridLock.Lock()
	defer pf.gridLock.Unlock()

	// Regenerate grid if teleport status changed
	var grid *game.Grid
	if pf.lastGrid != nil && pf.lastGridArea == a.Area && pf.lastTeleport == teleportEnabled {
		log.Printf("DEBUG: Using cached grid for area %d (teleport: %t)", a.Area, teleportEnabled)
		grid = pf.lastGrid
	} else {
		log.Printf("DEBUG: Regenerating grid for area %d (teleport: %t)", a.Area, teleportEnabled)
		grid = a.Grid.Copy()
		pf.preprocessGrid(grid, teleportEnabled)
		pf.lastGrid = grid
		pf.lastGridArea = a.Area
		pf.lastTeleport = teleportEnabled
	}

	// Handle cross-area pathing using grids & teleport status
	if !a.IsInside(to) {
		log.Printf("DEBUG: Handling cross-area pathing to %v", to)
		expandedGrid, err := pf.mergeGrids(to)
		if err != nil {
			log.Printf("ERROR: Failed to merge grids: %v", err)
			return nil, 0, false
		}
		pf.preprocessGrid(expandedGrid, teleportEnabled)
		grid = expandedGrid
		pf.lastGrid = grid
		pf.lastGridArea = a.Area
	}

	from = grid.RelativePosition(from)
	to = grid.RelativePosition(to)

	// Validate positions are within grid bounds
	if from.X < 0 || from.X >= grid.Width || from.Y < 0 || from.Y >= grid.Height ||
		to.X < 0 || to.X >= grid.Width || to.Y < 0 || to.Y >= grid.Height {
		log.Printf("WARN: Positions out of grid bounds (from: %v, to: %v)", from, to)
		return nil, 0, false
	}

	path, _, found := astar.CalculatePath(grid, a.Area, from, to, teleportEnabled)
	log.Printf("DEBUG: Path calculation result (found: %t, length: %d)", found, len(path))

	if config.Koolo.Debug.RenderMap {
		pf.renderMap(grid, from, to, path)
	}

	return path, len(path), found
}

// Enhanced grid preprocessing with teleport optimizations
func (pf *PathFinder) preprocessGrid(grid *game.Grid, teleportEnabled bool) {
	a := pf.data.AreaData
	log.Printf("DEBUG: Preprocessing grid for area %d (teleport: %t)", a.Area, teleportEnabled)

	// Teleport pathing optimizations
	// if teleportEnabled {
	// 	for y := 0; y < len(grid.CollisionGrid); y++ {
	// 		for x := 0; x < len(grid.CollisionGrid[y]); x++ {
	// 			// Only allow low priority for objects, keep walls blocked
	// 			switch grid.CollisionGrid[y][x] {
	// 			case game.CollisionTypeObject:
	// 				grid.CollisionGrid[y][x] = game.CollisionTypeLowPriority
	// 			}
	// 		}
	// 	}
	// }

	// Special area handling
	if a.Area == area.LutGholein {
		log.Printf("DEBUG: Applying Lut Gholein collision override")
		grid.CollisionGrid[13][210] = game.CollisionTypeNonWalkable
	}

	// Dynamic obstacle handling
	log.Printf("DEBUG: Processing %d objects", len(pf.data.AreaData.Objects))
	for _, o := range pf.data.AreaData.Objects {
		if o.IsChest() {
			log.Printf("DEBUG: Marking chest area around %v as non-walkable", o.Position)
			relativePos := grid.RelativePosition(o.Position)
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					y := relativePos.Y + dy
					x := relativePos.X + dx
					if y >= 0 && y < len(grid.CollisionGrid) && x >= 0 && x < len(grid.CollisionGrid[y]) {
						grid.CollisionGrid[y][x] = game.CollisionTypeNonWalkable
					}
				}
			}
			continue
		}

		if !grid.IsWalkable(o.Position) {
			log.Printf("DEBUG: Skipping non-walkable object %d at %v", o.Name, o.Position)
			continue
		}

		relativePos := grid.RelativePosition(o.Position)
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

	log.Printf("DEBUG: Processing %d monsters", len(pf.data.Monsters))
	for _, m := range pf.data.Monsters {
		if !grid.IsWalkable(m.Position) {
			log.Printf("DEBUG: Skipping non-walkable monster %d at %d:%d",
				m.Name,
				m.Position.X,
				m.Position.Y,
			)
			continue
		}
		relativePos := grid.RelativePosition(m.Position)
		grid.CollisionGrid[relativePos.Y][relativePos.X] = game.CollisionTypeMonster
	}
}

// Combine adjacent level grids for cross-area pathfinding
func (pf *PathFinder) mergeGrids(to data.Position) (*game.Grid, error) {
	log.Printf("DEBUG: Merging grids for position %v", to)
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
			log.Printf("DEBUG: Created merged grid size %dx%d", width, height)
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

// Enhanced path recovery with teleport adjustments
func (pf *PathFinder) GetClosestWalkablePath(dest data.Position) (Path, int, bool) {
	return pf.GetClosestWalkablePathFrom(pf.data.PlayerUnit.Position, dest)
}

// Find nearest accessible position when direct path is blocked
func (pf *PathFinder) GetClosestWalkablePathFrom(from, dest data.Position) (Path, int, bool) {
	a := pf.data.AreaData
	teleportEnabled := pf.data.CanTeleport()
	// First try direct path if destination is walkable or outside known area
	if a.IsWalkable(dest) || !a.IsInside(dest) {
		path, distance, found := pf.GetPath(dest)
		if found {
			return path, distance, found
		}
	}

	// Search in expanding squares around target position
	maxRange := 25
	step := 5
	if teleportEnabled {
		maxRange = 40
		step = 8
	}

	for dst := 1; dst < maxRange; dst += step {
		for i := -dst; i < dst; i += 1 {
			for j := -dst; j < dst; j += 1 {
				// Check perimeter of current search radius
				if math.Abs(float64(i)) >= math.Abs(float64(dst)) || math.Abs(float64(j)) >= math.Abs(float64(dst)) {
					cgY := dest.Y - pf.data.AreaOrigin.Y + j
					cgX := dest.X - pf.data.AreaOrigin.X + i
					if cgX > 0 && cgY > 0 && a.Height > cgY && a.Width > cgX {
						collisionType := a.CollisionGrid[cgY][cgX]
						// Adjust collision checks for teleport
						if teleportEnabled && collisionType == game.CollisionTypeLowPriority {
							return pf.GetPathFrom(from, data.Position{
								X: dest.X + i,
								Y: dest.Y + j,
							})
						}
						if collisionType == game.CollisionTypeWalkable {
							return pf.GetPathFrom(from, data.Position{
								X: dest.X + i,
								Y: dest.Y + j,
							})
						}
					}
				}
			}
		}
	}

	return nil, 0, false
}
