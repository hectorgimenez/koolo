package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"go.uber.org/zap"
)

type Factory struct {
	builder func(d data.Data) Action
	action  Action
	skip    bool
}

func NewFactory(builder func(d data.Data) Action) *Factory {
	return &Factory{
		builder: builder,
	}
}

func (f *Factory) NextStep(logger *zap.Logger, d data.Data) error {
	if f.skip {
		return ErrNoMoreSteps
	}

	if f.action == nil {
		f.action = f.builder(d)
		if f.action == nil {
			f.Skip()
			return ErrNoMoreSteps
		}
	}

	err := f.action.NextStep(logger, d)
	if errors.Is(err, ErrNoMoreSteps) {
		f.action = nil
		return nil
	}
	if errors.Is(err, ErrLogAndContinue) {
		return nil
	}

	return err
}

func (f *Factory) Skip() {
	f.skip = true
}
