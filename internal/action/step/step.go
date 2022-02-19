package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

const (
	StatusNotStarted Status = "NotStarted"
	StatusInProgress Status = "InProgress"
	StatusCompleted  Status = "Completed"
)

type Status string
type Step interface {
	Status(game.Data) Status
	Run(game.Data) error
}

type basicStep struct {
	status     Status
	lastChange time.Time
	lastRun    time.Time
}

func newBasicStep() basicStep {
	return basicStep{
		status: StatusNotStarted,
	}
}

func (bs *basicStep) tryTransitionStatus(to Status) Status {
	if bs.status == StatusCompleted {
		return StatusCompleted
	}
	if bs.status == StatusInProgress && to != StatusCompleted {
		return StatusInProgress
	}

	bs.status = to
	bs.lastChange = time.Now()
	return to
}
