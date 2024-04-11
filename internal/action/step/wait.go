package step

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
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

func (o *WaitStep) Status(_ game.Data, _ container.Container) Status {
	if o.status == StatusCompleted {
		return StatusCompleted
	}

	return o.status
}

func (o *WaitStep) Run(_ game.Data, _ container.Container) error {
	if o.firstRun.IsZero() {
		o.firstRun = time.Now()
		o.tryTransitionStatus(StatusInProgress)
	}

	if time.Since(o.firstRun) >= o.waitTime {
		o.tryTransitionStatus(StatusCompleted)
	}

	return nil
}
