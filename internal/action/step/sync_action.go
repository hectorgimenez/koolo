package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type SyncActionStep struct {
	basicStep
	action func(game.Data) error
}

func SyncAction(fn func(game.Data) error) *SyncActionStep {
	return &SyncActionStep{
		basicStep: newBasicStep(),
		action:    fn,
	}
}

func (s *SyncActionStep) Status(_ game.Data) Status {
	return s.status
}

func (s *SyncActionStep) Run(d game.Data) error {
	s.tryTransitionStatus(StatusCompleted)

	return s.action(d)
}
