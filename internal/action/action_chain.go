package action

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Chain struct {
	basicAction
	builder func(d game.Data) []Action
	actions []Action
}

func NewChain(builder func(d game.Data) []Action, opts ...Option) *Chain {
	a := &Chain{
		builder: builder,
	}

	for _, opt := range opts {
		opt(&a.basicAction)
	}

	return a
}

func (a *Chain) NextStep(d game.Data, container container.Container) error {
	if a.markSkipped {
		return ErrNoMoreSteps
	}

	if a.actions == nil {
		a.actions = a.builder(d)

		if a.abortOtherActionsIfNil {

			return errors.New(string(a.abortReason))
		}

		if a.actions == nil || len(a.actions) == 0 {
			a.Skip()
			return ErrNoMoreSteps
		}
	}

	var err error
	for _, action := range a.actions {
		err = action.NextStep(d, container)
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
