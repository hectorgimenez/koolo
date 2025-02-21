package pather

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Path []data.Position

var (
	// Spatial grid size for position normalization (5x5 tiles)
	gridSize = 5

	// Cache with timestamp-based eviction
	pathCache    = make(map[string]cacheEntry)
	cacheOrder   []string     // Maintains access order for LRU eviction
	cacheLock    sync.RWMutex // Protects concurrent access to cache
	maxCacheSize = 500        // Maximum number of paths to cache (increased from 100 for adaptive eviction)
)

type cacheEntry struct {
	path     Path
	lastUsed int64 // Unix nano timestamp
}

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
func getCachedPath(from, to data.Position, area area.ID, teleport bool) (Path, bool) {
	key := cacheKey(from, to, area, teleport)

	cacheLock.RLock()
	defer cacheLock.RUnlock()

	// Check direct path
	if entry, exists := pathCache[key]; exists {
		log.Printf("DEBUG: Path cache hit [direct] %s", key)
		return entry.path, true
	}

	// Check reverse path
	reverseKey := cacheKey(to, from, area, teleport)
	if entry, exists := pathCache[reverseKey]; exists {
		log.Printf("DEBUG: Path cache hit [reverse] %s", reverseKey)
		// Return reversed path if available
		reversed := make(Path, len(entry.path))
		for i, p := range entry.path {
			reversed[len(entry.path)-1-i] = p
		}
		return reversed, true
	}

	log.Printf("DEBUG: Path cache miss %s", key)
	return nil, false
}

// Stores path with usage timestamp
func cachePath(from, to data.Position, area area.ID, teleport bool, path Path) {
	key := cacheKey(from, to, area, teleport)
	log.Printf("DEBUG: Caching path %s (length: %d)", key, len(path))

	cacheLock.Lock()
	defer cacheLock.Unlock()

	// Update existing entry or create new
	pathCache[key] = cacheEntry{
		path:     path,
		lastUsed: time.Now().UnixNano(),
	}

	// Adaptive eviction when exceeding capacity
	if len(pathCache) >= maxCacheSize { // LRU eviction logic
		log.Printf("DEBUG: Performing cache eviction (current size: %d)", len(pathCache))
		var oldestKey string
		var oldestTime int64 = math.MaxInt64

		// Find least recently used entry
		for k, v := range pathCache {
			if v.lastUsed < oldestTime {
				oldestTime = v.lastUsed
				oldestKey = k
			}
		}
		delete(pathCache, oldestKey)
		log.Printf("DEBUG: Evicted cache entry %s", oldestKey)
	}
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
	log.Printf("DEBUG: Checking path intersection at %d:%d (padding: %d)", position.X, position.Y, padding)
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
