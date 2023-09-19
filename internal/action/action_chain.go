package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"go.uber.org/zap"
)

type Chain struct {
	basicAction
	builder func(d data.Data) []Action
	actions []Action
}

func NewChain(builder func(d data.Data) []Action, opts ...Option) *Chain {
	a := &Chain{
		builder: builder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *Chain) NextStep(logger *zap.Logger, d data.Data) error {
	if a.markSkipped {
		return ErrNoMoreSteps
	}

	if a.actions == nil {
		a.actions = a.builder(d)
		if a.actions == nil || len(a.actions) == 0 {
			a.Skip()
			return ErrNoMoreSteps
		}
	}

	var err error
	for _, action := range a.actions {
		err = action.NextStep(logger, d)
		if errors.Is(err, ErrNoMoreSteps) || errors.Is(err, ErrLogAndContinue) {
			continue
		}

		return err
	}

	// Reset actions, next iteration will try to build them again, if empty it will skip
	if (errors.Is(err, ErrNoMoreSteps) || errors.Is(err, ErrLogAndContinue)) && a.repeatUntilNoMoreSteps {
		a.actions = nil
		return nil
	}

	return ErrNoMoreSteps
}
