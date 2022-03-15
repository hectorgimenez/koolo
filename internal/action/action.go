package action

import (
	"errors"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const maxRetries = 5

var ErrWillBeRetried = errors.New("error occurred, but it will be retried")
var ErrNoRecover = errors.New("unrecoverable error occurred, game can not continue")
var ErrCanBeSkipped = errors.New("error occurred, but this action is not critical and game can continue")
var ErrNoMoreSteps = errors.New("action finished, no more steps remaining")

type Action interface {
	NextStep(data game.Data) error
	Skip()
}

type BasicAction struct {
	Steps             []step.Step
	builder           func(data game.Data) []step.Step
	builderExecuted   bool
	retries           int
	canBeSkipped      bool
	resetStepsOnError bool
	markSkipped       bool
}

func BuildOnRuntime(builder func(data game.Data) []step.Step, opts ...Option) *BasicAction {
	a := &BasicAction{
		builder: builder,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

type Option func(action *BasicAction)

func CanBeSkipped() Option {
	return func(action *BasicAction) {
		action.canBeSkipped = true
	}
}

func Resettable() Option {
	return func(action *BasicAction) {
		action.resetStepsOnError = true
	}
}

func (b *BasicAction) resetSteps() {
	if !b.resetStepsOnError {
		return
	}

	for _, s := range b.Steps {
		s.Reset()
	}
}

func (b *BasicAction) Skip() {
	if b.retries >= maxRetries && b.canBeSkipped {
		b.markSkipped = true
	}
}

func (b *BasicAction) NextStep(data game.Data) error {
	if b.markSkipped {
		return ErrNoMoreSteps
	}

	if b.retries >= maxRetries {
		if b.canBeSkipped {
			return fmt.Errorf("%w: attempt limit reached", ErrCanBeSkipped)
		}
		return fmt.Errorf("%w: attempt limit reached", ErrNoRecover)
	}

	if b.builder != nil && !b.builderExecuted {
		b.Steps = b.builder(data)
		b.builderExecuted = true
	}

	for _, s := range b.Steps {
		if s.Status(data) != step.StatusCompleted {
			err := s.Run(data)
			if err != nil {
				b.retries++
				b.resetSteps()
			}

			return nil
		}
	}

	return ErrNoMoreSteps
}
