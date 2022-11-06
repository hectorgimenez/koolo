package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"go.uber.org/zap"
)

type StaticAction struct {
	basicAction
	Steps           []step.Step
	builder         func(data game.Data) []step.Step
	builderExecuted bool
}

func BuildStatic(builder func(data game.Data) []step.Step, opts ...Option) *StaticAction {
	a := &StaticAction{
		builder: builder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *StaticAction) NextStep(logger *zap.Logger, data game.Data) error {
	if a.markSkipped {
		return ErrNoMoreSteps
	}

	if a.retries >= maxRetries {
		if a.canBeSkipped {
			return fmt.Errorf("%w: attempt limit reached", ErrCanBeSkipped)
		}
		return fmt.Errorf("%w: attempt limit reached", ErrNoRecover)
	}

	if a.builder != nil && !a.builderExecuted {
		a.Steps = a.builder(data)
		a.builderExecuted = true
	}

	for _, s := range a.Steps {
		if s.Status(data) != step.StatusCompleted {
			lastRun := s.LastRun()
			err := s.Run(data)
			if s.LastRun().After(lastRun) {
				//logger.Debug("Executed step", zap.String("step_name", reflect.TypeOf(s).Elem().Name()))
			}
			if err != nil {
				a.retries++
				a.resetSteps()
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
