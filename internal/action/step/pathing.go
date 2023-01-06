package step

import (
	"errors"
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	maxPathNotFoundRetries = 5
)

var errPathNotFound = errors.New("path not found")

type pathingStep struct {
	basicStep
	consecutivePathNotFound int
	path                    []astar.Pather
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

func (s *pathingStep) cachePath(data game.Data) bool {
	nearestKey := 0
	nearestDistance := 99999999
	for k, pos := range s.path {
		distance := pather.DistanceFromMe(data, pos.(*pather.Tile).X+data.AreaOrigin.X, pos.(*pather.Tile).Y+data.AreaOrigin.Y)
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

func (s *pathingStep) isPlayerStuck(data game.Data) bool {
	s.lastRunPositions = append(s.lastRunPositions, [2]int{data.PlayerUnit.Position.X, data.PlayerUnit.Position.Y})
	if len(s.lastRunPositions) > 20 {
		s.lastRunPositions = s.lastRunPositions[1:]
	} else {
		return false
	}

	for _, pos := range s.lastRunPositions {
		if pos[0] != data.PlayerUnit.Position.X || pos[1] != data.PlayerUnit.Position.Y {
			return false
		}
	}

	return true
}
