package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"go.uber.org/zap"
)

type Chain struct {
	builder func(d data.Data) []Action
	actions []Action
	skip    bool
}

func NewChain(builder func(d data.Data) []Action) *Chain {
	return &Chain{
		builder: builder,
	}
}

func (a *Chain) NextStep(logger *zap.Logger, d data.Data) error {
	if a.skip {
		return ErrNoMoreSteps
	}

	if a.actions == nil {
		a.actions = a.builder(d)
		if a.actions == nil {
			a.Skip()
			return ErrNoMoreSteps
		}
	}

	for _, action := range a.actions {
		err := action.NextStep(logger, d)
		if errors.Is(err, ErrNoMoreSteps) || errors.Is(err, ErrLogAndContinue) {
			continue
		}

		return err
	}

	return ErrNoMoreSteps
}

func (a *Chain) Skip() {
	a.skip = true
}
