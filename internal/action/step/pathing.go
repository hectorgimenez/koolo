package step

import (
	"errors"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	maxPathNotFoundRetries = 5
)

var errPathNotFound = errors.New("path not found")

type pathingStep struct {
	basicStep
	consecutivePathNotFound int
	path                    *pather.Pather
	lastRunPositions        [][2]int
	blacklistedPositions    [][2]int
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
	for k, pos := range s.path.AstarPather {
		destination := data.Position{
			X: pos.(*pather.Tile).X + d.AreaOrigin.X,
			Y: pos.(*pather.Tile).Y + d.AreaOrigin.Y,
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
