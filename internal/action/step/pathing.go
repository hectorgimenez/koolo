package step

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type pathingStep struct {
	basicStep
	consecutivePathNotFound int
	path                    pather.Pather
	lastRunPositions        []data.Position
	blacklistedPositions    [][2]int
	previousArea            area.ID
}

func newPathingStep() pathingStep {
	return pathingStep{
		basicStep: basicStep{
			status: StatusNotStarted,
		},
	}
}

func (s *pathingStep) cachePath(d game.Data) bool {
	nearestKey := 0
	nearestDistance := 99999999

	if s.previousArea != d.PlayerUnit.Area {
		return false
	}

	for k, pos := range s.path {
		tile := pos.(*pather.Tile)
		expandedGrid := 0
		if len(tile.W.World) == 3000 {
			expandedGrid = 1500
		}

		destination := data.Position{
			X: tile.X + (d.AreaOrigin.X - expandedGrid),
			Y: tile.Y + (d.AreaOrigin.Y - expandedGrid),
		}

		distance := pather.DistanceFromMe(d, destination)
		if distance < nearestDistance {
			nearestDistance = distance
			nearestKey = k
		}
	}

	if nearestDistance < 5 && len(s.path) > nearestKey {
		//fmt.Println(fmt.Sprintf("Max deviation: %d, using Path Key: %d [%d]", nearestDistance, nearestKey, len(s.path)-1))
		s.path = s.path[:nearestKey]

		return true
	}

	return false
}

func (s *pathingStep) isPlayerStuck(d game.Data) bool {
	if s.lastRun.IsZero() {
		return false
	}

	if len(s.lastRunPositions) > 3 {
		s.lastRunPositions = s.lastRunPositions[1:]
	} else {
		return false
	}

	for _, pos := range s.lastRunPositions {
		if pather.DistanceFromMe(d, pos) > 5 {
			return false
		}
	}

	return true
}
