package action

import (
	"fmt"
	"reflect"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"go.uber.org/zap"
)

type DynamicAction struct {
	basicAction
	stepBuilder func(d data.Data) ([]step.Step, bool)
	steps       []step.Step
	finished    bool
}

func BuildDynamic(stepBuilder func(d data.Data) ([]step.Step, bool), opts ...Option) *DynamicAction {
	a := &DynamicAction{
		stepBuilder: stepBuilder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *DynamicAction) NextStep(logger *zap.Logger, d data.Data) error {
	if a.markSkipped || a.finished {
		return ErrNoMoreSteps
	}

	if a.steps == nil {
		steps, ok := a.stepBuilder(d)
		if !ok {
			a.finished = true
			return ErrNoMoreSteps
		}
		a.steps = steps
	}

	for _, s := range a.steps {
		if s.Status(d) != step.StatusCompleted {
			err := s.Run(d)
			if err != nil {
				a.retries++
			}

			if a.retries >= maxRetries {
				if a.canBeSkipped {
					return fmt.Errorf("%w: attempt limit reached on step: %s: %v", ErrCanBeSkipped, reflect.TypeOf(s).Elem().Name(), err)
				}
				return fmt.Errorf("%w: attempt limit reached on step: %s: %v", ErrNoRecover, reflect.TypeOf(s).Elem().Name(), err)
			}

			return nil
		}
	}

	steps, ok := a.stepBuilder(d)
	if !ok {
		a.finished = true
		return ErrNoMoreSteps
	}

	a.steps = steps

	return nil
}
