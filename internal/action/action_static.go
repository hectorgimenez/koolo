package action

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"go.uber.org/zap"
)

type StaticAction struct {
	basicAction
	Steps           []step.Step
	builder         func(d data.Data) []step.Step
	builderExecuted bool
}

func BuildStatic(builder func(d data.Data) []step.Step, opts ...Option) *StaticAction {
	a := &StaticAction{
		builder: builder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *StaticAction) NextStep(logger *zap.Logger, d data.Data) error {
	if a.markSkipped {
		return ErrNoMoreSteps
	}

	if a.builder != nil && !a.builderExecuted {
		a.Steps = a.builder(d)
		a.builderExecuted = true
	}

	for _, s := range a.Steps {
		if s.Status(d) != step.StatusCompleted {
			lastRun := s.LastRun()
			err := s.Run(d)
			if s.LastRun().After(lastRun) {
				//logger.Debug("Executed step", zap.String("step_name", reflect.TypeOf(s).Elem().Name()))
			}
			if err != nil {
				a.retries++
				a.resetSteps()
			}

			if a.retries >= maxRetries {
				if a.ignoreErrors {
					a.markSkipped = true
					return errors.Join(err, ErrLogAndContinue)
				}
				if a.canBeSkipped {
					return fmt.Errorf("%w: attempt limit reached on step: %s: %v", ErrCanBeSkipped, reflect.TypeOf(s).Elem().Name(), err)
				}
				return fmt.Errorf("%w: attempt limit reached on step: %s: %v", ErrNoRecover, reflect.TypeOf(s).Elem().Name(), err)
			}

			return nil
		}
	}

	return ErrNoMoreSteps
}

func (a *StaticAction) resetSteps() {
	if !a.resetStepsOnError {
		return
	}

	for _, s := range a.Steps {
		s.Reset()
	}
}
