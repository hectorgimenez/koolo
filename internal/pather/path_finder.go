package pather

import (
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather/astar"
)

const (
	positionRounding = 5                // Round positions to nearest 5 units
	maxCacheSize     = 10000            // Increased from 2000 to handle longer sessions
	cacheExpiration  = 30 * time.Minute // Automatically clear old entries
)

// cacheKey includes game seed and area to avoid stale paths across games
type cacheKey struct {
	fromX, fromY int // Rounded start position
	toX, toY     int // Rounded end position
	area         area.ID
	mapSeed      uint // Unique per game session
}

// cacheEntry stores the cached path result
type cacheEntry struct {
	path     Path
	distance int
	found    bool
	lastUsed time.Time // Track usage for LRU
}

type PathFinder struct {
	gr          *game.MemoryReader
	data        *game.Data
	hid         *game.HID
	cfg         *config.CharacterCfg
	cache       map[cacheKey]cacheEntry
	gridCache   map[area.ID]*game.Grid // Cached grids per area
	currentSeed uint                   // Track current game seed
	mu          sync.Mutex
	cacheHits   int // Track cache Hits
	cacheMisses int // Track cache Misses
}

func NewPathFinder(gr *game.MemoryReader, data *game.Data, hid *game.HID, cfg *config.CharacterCfg) *PathFinder {
	pf := &PathFinder{
		gr:        gr,
		data:      data,
		hid:       hid,
		cfg:       cfg,
		cache:     make(map[cacheKey]cacheEntry),
		gridCache: make(map[area.ID]*game.Grid),
	}
	go pf.logCacheStats()
	return pf
}

// GetPath attempts to find a path from the player's current position to the target
// First tries a direct path, then falls back to finding a nearby walkable position
func (pf *PathFinder) GetPath(to data.Position) (Path, int, bool) {
	// First try direct path
	if path, distance, found := pf.GetPathFrom(pf.data.PlayerUnit.Position, to); found {
		return path, distance, true
	}

	// If direct path fails, try to find nearby walkable position
	if walkableTo, found := pf.findNearbyWalkablePosition(to); found {
		return pf.GetPathFrom(pf.data.PlayerUnit.Position, walkableTo)
	}

	return nil, 0, false
}

// GetPathFrom calculates a path between two positions, using cached results when available
func (pf *PathFinder) GetPathFrom(from, to data.Position) (Path, int, bool) {
	currentArea := pf.data.PlayerUnit.Area
	currentSeed := pf.gr.MapSeed()

	// Clear cache when new game is detected
	pf.mu.Lock()
	if pf.currentSeed != currentSeed {
		log.Printf(
			"Clearing cache - Old Seed: %d | New Seed: %d | Cache Size Before: %d | Grid Cache Before: %d",
			pf.currentSeed,
			currentSeed,
			len(pf.cache),
			len(pf.gridCache),
		)
		pf.cache = make(map[cacheKey]cacheEntry)
		pf.gridCache = make(map[area.ID]*game.Grid)
		pf.currentSeed = currentSeed
	}
	pf.mu.Unlock()

	// Round positions to reduce key fragmentation
	roundedFromX := from.X / positionRounding
	roundedFromY := from.Y / positionRounding
	roundedToX := to.X / positionRounding
	roundedToY := to.Y / positionRounding

	key := cacheKey{
		fromX:   roundedFromX,
		fromY:   roundedFromY,
		toX:     roundedToX,
		toY:     roundedToY,
		area:    currentArea,
		mapSeed: currentSeed,
	}

	// Cache check
	pf.mu.Lock()
	entry, cached := pf.cache[key]
	if cached {
		// Update last used time
		entry.lastUsed = time.Now()
		pf.cache[key] = entry
		pf.cacheHits++
		pf.mu.Unlock()
		return entry.path, entry.distance, entry.found
	}
	pf.mu.Unlock()

	pf.mu.Lock()
	pf.cacheMisses++
	pf.mu.Unlock()

	a := pf.data.AreaData

	// Optimized grid handling: reuse cached grid if available
	pf.mu.Lock()
	grid, ok := pf.gridCache[a.Area]
	pf.mu.Unlock()
	if !ok || currentArea == area.ArcaneSanctuary { // Always recopy dynamic areas
		grid = a.Grid.Copy()
		pf.mu.Lock()
		pf.gridCache[a.Area] = grid
		pf.mu.Unlock()
	}

	// Special handling for Arcane Sanctuary (to allow pathing with platforms)
	if currentArea == area.ArcaneSanctuary && pf.data.CanTeleport() {
		// Make all non-walkable tiles into low priority tiles for teleport pathing
		for y := 0; y < len(grid.CollisionGrid); y++ {
			for x := 0; x < len(grid.CollisionGrid[y]); x++ {
				if grid.CollisionGrid[y][x] == game.CollisionTypeNonWalkable {
					grid.CollisionGrid[y][x] = game.CollisionTypeLowPriority
				}
			}
		}
	}
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

	// Cache result ONLY IF FOUND
	if found {
		pf.mu.Lock()
		// Perform maintenance before adding new entry (already locked)
		now := time.Now()
		for k, v := range pf.cache {
			if now.Sub(v.lastUsed) > cacheExpiration {
				delete(pf.cache, k)
			}
		}

		// Remove oldest entries if still over capacity
		if len(pf.cache) >= maxCacheSize {
			type entry struct {
				key cacheKey
				t   time.Time
			}

			entries := make([]entry, 0, len(pf.cache))
			for k, v := range pf.cache {
				entries = append(entries, entry{k, v.lastUsed})
			}

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].t.Before(entries[j].t)
			})

			toRemove := len(entries) - maxCacheSize
			if toRemove > 0 {
				for _, e := range entries[:toRemove] {
					delete(pf.cache, e.key)
				}
			}
		}

		// Store the new path
		pf.cache[key] = cacheEntry{
			path:     path,
			distance: distance,
			found:    found,
			lastUsed: time.Now(),
		}
		pf.mu.Unlock()
	}

	return path, distance, found
}

// DEBUG ONLY
// logCacheStats prints cache metrics every 10 seconds
func (pf *PathFinder) logCacheStats() {
	for {
		pf.mu.Lock()
		// Perform maintenance while locked
		now := time.Now()
		for k, v := range pf.cache {
			if now.Sub(v.lastUsed) > cacheExpiration {
				delete(pf.cache, k)
			}
		}

		currentSeed := pf.gr.MapSeed()
		if currentSeed == 0 {
			pf.mu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf(
			"Path Cache - Hits: %d | Misses: %d | Size: %d | Grid Cache Size: %d | mapSeed: %d",
			pf.cacheHits,
			pf.cacheMisses,
			len(pf.cache),
			len(pf.gridCache),
			currentSeed,
		)
		pf.mu.Unlock()
		time.Sleep(10 * time.Second)
	}
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

func (pf *PathFinder) GetClosestWalkablePath(dest data.Position) (Path, int, bool) {
	return pf.GetClosestWalkablePathFrom(pf.data.PlayerUnit.Position, dest)
}

func (pf *PathFinder) GetClosestWalkablePathFrom(from, dest data.Position) (Path, int, bool) {
	a := pf.data.AreaData
	if a.IsWalkable(dest) || !a.IsInside(dest) {
		path, distance, found := pf.GetPath(dest)
		if found {
			return path, distance, found
		}
	}

	maxRange := 20
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

func (pf *PathFinder) findNearbyWalkablePosition(target data.Position) (data.Position, bool) {
	// Search in expanding squares around the target position
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
