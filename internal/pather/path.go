package pather

import (
	"fmt"
	"sync"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Path []data.Position

var (
	pathCache    = make(map[string]Path)
	cacheOrder   []string
	cacheLock    sync.RWMutex
	maxCacheSize = 100
)

func cacheKey(from, to data.Position, area area.ID) string {
	return fmt.Sprintf("%d:%d:%d:%d:%d", from.X, from.Y, to.X, to.Y, area)
}

func getCachedPath(from, to data.Position, area area.ID) (Path, bool) {
	key := cacheKey(from, to, area)
	cacheLock.RLock()
	defer cacheLock.RUnlock()
	return pathCache[key], pathCache[key] != nil
}

func cachePath(from, to data.Position, area area.ID, path Path) {
	key := cacheKey(from, to, area)
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if len(pathCache) >= maxCacheSize {
		evictKey := cacheOrder[len(cacheOrder)-1]
		delete(pathCache, evictKey)
		cacheOrder = cacheOrder[:len(cacheOrder)-1]
	}

	pathCache[key] = path
	cacheOrder = append([]string{key}, cacheOrder...)
}

func (p Path) To() data.Position {
	if len(p) == 0 {
		return data.Position{}
	}
	return data.Position{
		X: p[len(p)-1].X,
		Y: p[len(p)-1].Y,
	}
}

func (p Path) From() data.Position {
	if len(p) == 0 {
		return data.Position{}
	}
	return data.Position{
		X: p[0].X,
		Y: p[0].Y,
	}
}

func (p Path) Intersects(d game.Data, position data.Position, padding int) bool {
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
