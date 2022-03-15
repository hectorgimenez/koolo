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
	Reset()
}

type basicStep struct {
	status  Status
	lastRun time.Time
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
	return to
}

func (bs *basicStep) Reset() {
	bs.status = StatusNotStarted
	bs.lastRun = time.Time{}
}
