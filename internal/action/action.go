package action

import (
	"errors"
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

type basicAction struct {
	retries           int
	canBeSkipped      bool
	resetStepsOnError bool
	markSkipped       bool
}

type Option func(action *basicAction)

func CanBeSkipped() Option {
	return func(action *basicAction) {
		action.canBeSkipped = true
	}
}

func Resettable() Option {
	return func(action *basicAction) {
		action.resetStepsOnError = true
	}
}

func (b *basicAction) Skip() {
	if b.retries >= maxRetries && b.canBeSkipped {
		b.markSkipped = true
	}
}
