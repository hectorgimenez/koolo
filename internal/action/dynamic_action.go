package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"go.uber.org/zap"
	"reflect"
)

type DynamicAction struct {
	basicAction
	stepBuilder func(data game.Data) (step.Step, bool)
	currentStep step.Step
}

func BuildDynamic(stepBuilder func(data game.Data) (step.Step, bool), opts ...Option) *DynamicAction {
	a := &DynamicAction{
		stepBuilder: stepBuilder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *DynamicAction) NextStep(logger *zap.Logger, data game.Data) error {
	if a.markSkipped {
		return ErrNoMoreSteps
	}

	if a.retries >= maxRetries {
		if a.canBeSkipped {
			return fmt.Errorf("%w: attempt limit reached", ErrCanBeSkipped)
		}
		return fmt.Errorf("%w: attempt limit reached", ErrNoRecover)
	}

	if a.currentStep == nil {
		nextStep, ok := a.stepBuilder(data)
		if !ok {
			return ErrNoMoreSteps
		}
		a.currentStep = nextStep
	}

	if a.currentStep.Status(data) != step.StatusCompleted {
		logger.Debug("Running step", zap.String("step_name", reflect.TypeOf(a.currentStep).Elem().Name()))
		lastRun := a.currentStep.LastRun()
		err := a.currentStep.Run(data)
		if a.currentStep.LastRun().After(lastRun) {
			//logger.Debug("Executed step", zap.String("step_name", reflect.TypeOf(a.currentStep).Elem().Name()))
		}
		if err != nil {
			a.retries++
		}
	} else {
		a.currentStep = nil
	}

	return nil
}
