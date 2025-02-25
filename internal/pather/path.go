package pather

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Path []data.Position

// PathCacheEntry stores path data and metadata in the global cache
type PathCacheEntry struct {
	Path             Path
	LastUsed         int64 // Unix nano timestamp
	PreviousPosition data.Position
	LastCheck        time.Time
	LastRun          time.Time
}

var (
	// Spatial grid size for position normalization (5x5 tiles)
	gridSize = 5

	// Cache with timestamp-based eviction
	pathCache    = make(map[string]PathCacheEntry)
	cacheLock    sync.RWMutex // Protects concurrent access to cache
	maxCacheSize = 500        // Maximum number of paths to cache
)

// Generates normalized cache key using spatial partitioning
func cacheKey(from, to data.Position, area area.ID, teleport bool) string {
	// Quantize positions to grid cells to group similar paths
	normFromX := (from.X / gridSize) * gridSize
	normFromY := (from.Y / gridSize) * gridSize
	normToX := (to.X / gridSize) * gridSize
	normToY := (to.Y / gridSize) * gridSize

	return fmt.Sprintf("%d:%d:%d:%d:%d:%t",
		normFromX, normFromY,
		normToX, normToY,
		area, teleport)
}

// Retrieves cached path with reverse path check
func GetCachedPath(from, to data.Position, area area.ID, teleport bool) (Path, bool) {
	key := cacheKey(from, to, area, teleport)

	cacheLock.RLock()
	defer cacheLock.RUnlock()

	// Check direct path
	if entry, exists := pathCache[key]; exists {
		return entry.Path, true
	}

	// Check reverse path
	reverseKey := cacheKey(to, from, area, teleport)
	if entry, exists := pathCache[reverseKey]; exists {
		// Return reversed path if available
		reversed := make(Path, len(entry.Path))
		for i, p := range entry.Path {
			reversed[len(entry.Path)-1-i] = p
		}
		return reversed, true
	}

	return nil, false
}

// GetCacheEntry retrieves the full cache entry for a path
func GetCacheEntry(from, to data.Position, area area.ID, teleport bool) (*PathCacheEntry, bool) {
	key := cacheKey(from, to, area, teleport)

	cacheLock.RLock()
	defer cacheLock.RUnlock()

	if entry, exists := pathCache[key]; exists {
		entryCopy := entry // Create a copy to safely return
		return &entryCopy, true
	}

	return nil, false
}

// StorePath stores a path in the global cache
func StorePath(from, to data.Position, area area.ID, teleport bool, path Path, currentPos data.Position) {
	key := cacheKey(from, to, area, teleport)

	cacheLock.Lock()
	defer cacheLock.Unlock()

	entry := PathCacheEntry{
		Path:             path,
		LastUsed:         time.Now().UnixNano(),
		PreviousPosition: currentPos,
		LastCheck:        time.Now(),
		LastRun:          time.Time{},
	}

	pathCache[key] = entry

	// Adaptive eviction when exceeding capacity
	if len(pathCache) >= maxCacheSize { // LRU eviction logic
		var oldestKey string
		var oldestTime int64 = math.MaxInt64

		// Find least recently used entry
		for k, v := range pathCache {
			if v.LastUsed < oldestTime {
				oldestTime = v.LastUsed
				oldestKey = k
			}
		}
		delete(pathCache, oldestKey)
	}
}

// UpdatePathLastRun updates the lastRun time for a path
func UpdatePathLastRun(from, to data.Position, area area.ID, teleport bool) {
	key := cacheKey(from, to, area, teleport)

	cacheLock.Lock()
	defer cacheLock.Unlock()

	if entry, exists := pathCache[key]; exists {
		entry.LastRun = time.Now()
		entry.LastUsed = time.Now().UnixNano()
		pathCache[key] = entry
	}
}

// UpdatePathLastCheck updates the lastCheck time and previous position for a path
func UpdatePathLastCheck(from, to data.Position, area area.ID, teleport bool, currentPos data.Position) {
	key := cacheKey(from, to, area, teleport)

	cacheLock.Lock()
	defer cacheLock.Unlock()

	if entry, exists := pathCache[key]; exists {
		entry.LastCheck = time.Now()
		entry.PreviousPosition = currentPos
		entry.LastUsed = time.Now().UnixNano()
		pathCache[key] = entry
	}
}

// IsPathValid checks if a path is still valid based on current position
func IsPathValid(currentPos data.Position, startPos data.Position, destPos data.Position, path Path) bool {
	// Valid if we're close to start, destination, or current path
	if DistanceFromPoint(currentPos, startPos) < 20 ||
		DistanceFromPoint(currentPos, destPos) < 20 {
		return true
	}

	// Check if we're near any point on the path
	minDistance := 20
	for _, pathPoint := range path {
		dist := DistanceFromPoint(currentPos, pathPoint)
		if dist < minDistance {
			minDistance = dist
			break
		}
	}

	return minDistance < 20
}

// Return the ending position of the path
func (p Path) To() data.Position {
	if len(p) == 0 {
		return data.Position{}
	}
	return data.Position{
		X: p[len(p)-1].X,
		Y: p[len(p)-1].Y,
	}
}

// Return the starting position of the path
func (p Path) From() data.Position {
	if len(p) == 0 {
		return data.Position{}
	}
	return data.Position{
		X: p[0].X,
		Y: p[0].Y,
	}
}

// Intersects checks if the given position intersects with the path
func (p Path) Intersects(d game.Data, position data.Position, padding int) bool {
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
