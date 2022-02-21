package step

import "errors"

const (
	maxPathNotFoundRetries = 5
	maxPlayerStuckRetries  = 5
)

var errPathNotFound = errors.New("path not found")

type pathingStep struct {
	basicStep
	consecutivePathNotFound int
	consecutivePlayerStuck  int
}

func newPathingStep() pathingStep {
	return pathingStep{
		basicStep: basicStep{
			status: StatusNotStarted,
		},
	}
}
