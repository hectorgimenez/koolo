package action

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/game"
	"go.uber.org/zap"
)

type Factory struct {
	builder func(data game.Data) Action
	action  Action
	skip    bool
}

func NewFactory(builder func(data game.Data) Action) *Factory {
	return &Factory{
		builder: builder,
	}
}

func (f *Factory) NextStep(logger *zap.Logger, data game.Data) error {
	if f.skip {
		return ErrNoMoreSteps
	}

	if f.action == nil {
		f.action = f.builder(data)
		if f.action == nil {
			f.Skip()
			return ErrNoMoreSteps
		}
	}

	err := f.action.NextStep(logger, data)
	if errors.Is(err, ErrNoMoreSteps) {
		f.action = nil
		return nil
	}

	return err
}

func (f *Factory) Skip() {
	f.skip = true
}
