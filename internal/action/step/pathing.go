package step

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type pathingStep struct {
	basicStep
	consecutivePathNotFound int
	path                    *pather.Pather
	lastRunPositions        [][2]int
	blacklistedPositions    [][2]int
	previousArea            area.Area
}

func newPathingStep() pathingStep {
	return pathingStep{
		basicStep: basicStep{
			status: StatusNotStarted,
		},
	}
}

func (s *pathingStep) cachePath(d data.Data) bool {
	nearestKey := 0
	nearestDistance := 99999999

	if s.previousArea != d.PlayerUnit.Area {
		return false
	}

	for k, pos := range s.path.AstarPather {
		tile := pos.(*pather.Tile)
		expandedGrid := 0
		if len(tile.W) == 3000 {
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

	if nearestDistance < 5 && len(s.path.AstarPather) > nearestKey {
		//fmt.Println(fmt.Sprintf("Max deviation: %d, using Path Key: %d [%d]", nearestDistance, nearestKey, len(s.path)-1))
		s.path.AstarPather = s.path.AstarPather[:nearestKey]

		return true
	}

	return false
}

func (s *pathingStep) isPlayerStuck(d data.Data) bool {
	if s.lastRun.IsZero() {
		return false
	}

	s.lastRunPositions = append(s.lastRunPositions, [2]int{d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y})
	if len(s.lastRunPositions) > 20 {
		s.lastRunPositions = s.lastRunPositions[1:]
	} else {
		return false
	}

	for _, pos := range s.lastRunPositions {
		if pos[0] != d.PlayerUnit.Position.X || pos[1] != d.PlayerUnit.Position.Y {
			return false
		}
	}

	return true
}
