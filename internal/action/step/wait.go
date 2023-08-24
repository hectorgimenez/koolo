package step

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
)

type WaitStep struct {
	basicStep
	waitTime time.Duration
	firstRun time.Time
}

func Wait(duration time.Duration) *WaitStep {
	return &WaitStep{
		basicStep: newBasicStep(),
		waitTime:  duration,
	}
}

func (o *WaitStep) Status(_ data.Data) Status {
	if o.status == StatusCompleted {
		return StatusCompleted
	}

	return o.status
}

func (o *WaitStep) Run(_ data.Data) error {
	if o.firstRun.IsZero() {
		o.firstRun = time.Now()
		o.tryTransitionStatus(StatusInProgress)
	}

	if time.Since(o.firstRun) >= o.waitTime {
		o.tryTransitionStatus(StatusCompleted)
	}

	return nil
}
