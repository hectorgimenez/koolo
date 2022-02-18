package step

import (
	"github.com/hectorgimenez/koolo/internal/game"
)

type SyncAction struct {
	basicStep
	action func(game.Data) error
}

func NewSyncAction(fn func(game.Data) error) *SyncAction {
	return &SyncAction{
		basicStep: newBasicStep(),
		action:    fn,
	}
}

func (s *SyncAction) Status(_ game.Data) Status {
	return s.status
}

func (s *SyncAction) Run(d game.Data) error {
	s.tryTransitionStatus(StatusCompleted)

	return s.action(d)
}
