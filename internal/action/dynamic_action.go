package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"go.uber.org/zap"
)

type DynamicAction struct {
	basicAction
	stepBuilder func(data game.Data) ([]step.Step, bool)
	steps       []step.Step
	finished    bool
}

func BuildDynamic(stepBuilder func(data game.Data) ([]step.Step, bool), opts ...Option) *DynamicAction {
	a := &DynamicAction{
		stepBuilder: stepBuilder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *DynamicAction) NextStep(logger *zap.Logger, data game.Data) error {
	if a.markSkipped || a.finished {
		return ErrNoMoreSteps
	}

	if a.retries >= maxRetries {
		if a.canBeSkipped {
			return fmt.Errorf("%w: attempt limit reached", ErrCanBeSkipped)
		}
		return fmt.Errorf("%w: attempt limit reached", ErrNoRecover)
	}

	if a.steps == nil {
		steps, ok := a.stepBuilder(data)
		if !ok {
			return ErrNoMoreSteps
		}
		a.steps = steps
	}

	for _, s := range a.steps {
		if s.Status(data) != step.StatusCompleted {
			err := s.Run(data)
			if err != nil {
				a.retries++
			}

			return nil
		}
	}

	steps, ok := a.stepBuilder(data)
	if !ok {
		a.finished = true
		return ErrNoMoreSteps
	}

	a.steps = steps

	return nil
}
