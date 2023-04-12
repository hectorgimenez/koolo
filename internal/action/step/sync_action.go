package step

import "github.com/hectorgimenez/d2go/pkg/data"

type SyncActionStep struct {
	basicStep
	action        func(data.Data) error
	statusChecker func(data.Data) Status
}

func SyncStep(action func(data.Data) error) *SyncActionStep {
	return &SyncActionStep{
		basicStep: newBasicStep(),
		action:    action,
	}
}

func SyncStepWithCheck(action func(data.Data) error, statusChecker func(data.Data) Status) *SyncActionStep {
	return &SyncActionStep{
		basicStep:     newBasicStep(),
		action:        action,
		statusChecker: statusChecker,
	}
}

func (s *SyncActionStep) Status(d data.Data) Status {
	if s.status == StatusCompleted {
		return StatusCompleted
	}

	if s.statusChecker != nil {
		s.tryTransitionStatus(s.statusChecker(d))
	}

	return s.status
}

func (s *SyncActionStep) Run(d data.Data) error {
	if s.statusChecker == nil {
		s.tryTransitionStatus(StatusCompleted)
	}

	err := s.action(d)
	return err
}
