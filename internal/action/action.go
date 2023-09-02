package action

import (
	"errors"

	"github.com/hectorgimenez/d2go/pkg/data"
	"go.uber.org/zap"
)

const maxRetries = 5

var ErrWillBeRetried = errors.New("error occurred, but it will be retried")
var ErrNoRecover = errors.New("unrecoverable error occurred, game can not continue")
var ErrCanBeSkipped = errors.New("error occurred, but this action is not critical and game can continue")
var ErrNoMoreSteps = errors.New("action finished, no more steps remaining")
var ErrLogAndContinue = errors.New("error occurred, but marking action as completed")

type Action interface {
	NextStep(logger *zap.Logger, d data.Data) error
	Skip()
}

type basicAction struct {
	retries                int
	canBeSkipped           bool
	resetStepsOnError      bool
	markSkipped            bool
	ignoreErrors           bool
	repeatUntilNoMoreSteps bool
}

type Option func(action *basicAction)

func CanBeSkipped() Option {
	return func(action *basicAction) {
		action.canBeSkipped = true
	}
}

// IgnoreErrors will ignore the errors and mark the action as succeeded
func IgnoreErrors() Option {
	return func(action *basicAction) {
		action.ignoreErrors = true
	}
}

func Resettable() Option {
	return func(action *basicAction) {
		action.resetStepsOnError = true
	}
}

func RepeatUntilNoSteps() Option {
	return func(action *basicAction) {
		action.repeatUntilNoMoreSteps = true
	}
}

func (b *basicAction) Skip() {
	b.markSkipped = true
}
