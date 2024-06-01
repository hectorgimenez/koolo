package action

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/koolo/internal/action/step"
)

type StepChainAction struct {
	basicAction
	Steps           []step.Step
	builder         func(d game.Data) []step.Step
	builderExecuted bool
}

func NewStepChain(builder func(d game.Data) []step.Step, opts ...Option) *StepChainAction {
	a := &StepChainAction{
		builder: builder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *StepChainAction) NextStep(d game.Data, container container.Container) error {
	// Ensure that we first check if its nill to avoid accessing a nill value (can result in panic)
	if a == nil || a.isFinished {
		return ErrNoMoreSteps
	}

	if a.builder != nil && !a.builderExecuted {
		a.Steps = a.builder(d)
		a.builderExecuted = true
	}

	for _, s := range a.Steps {
		if s.Status(d, container) != step.StatusCompleted {
			lastRun := s.LastRun()
			err := s.Run(d, container)
			if s.LastRun().After(lastRun) {
				//logger.Debug("Executed step", slog.String("step_name", reflect.TypeOf(s).Elem().Name()))
			}
			if err != nil {
				a.retries++
				a.resetSteps()
			}

			if a.retries >= maxRetries {
				if a.ignoreErrors {
					a.isFinished = true
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

	if a.repeatUntilNoMoreSteps {
		a.Steps = a.builder(d)
		// Continue execution if builder still returns steps
		if a.Steps != nil && len(a.Steps) > 0 {
			return nil
		} else {
			a.isFinished = true
		}
	}

	a.isFinished = true

	return ErrNoMoreSteps
}

func (a *StepChainAction) resetSteps() {
	if !a.resetStepsOnError {
		return
	}

	for _, s := range a.Steps {
		s.Reset()
	}
}
