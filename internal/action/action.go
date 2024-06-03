package action

import (
	"errors"

	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
)

type AbortReason string

const (
	ReasonChicken AbortReason = "chicken occurred"
)

const maxRetries = 3

var ErrWillBeRetried = errors.New("error occurred, but it will be retried")
var ErrNoRecover = errors.New("unrecoverable error occurred, game can not continue")
var ErrCanBeSkipped = errors.New("error occurred, but this action is not critical and game can continue")
var ErrNoMoreSteps = errors.New("action finished, no more steps remaining")
var ErrLogAndContinue = errors.New("error occurred, but marking action as completed")

type Action interface {
	NextStep(d game.Data, container container.Container) error
	Skip()
	IsFinished() bool
}

type basicAction struct {
	retries                int
	canBeSkipped           bool
	resetStepsOnError      bool
	isFinished             bool
	ignoreErrors           bool
	repeatUntilNoMoreSteps bool
	abortOtherActionsIfNil bool
	abortReason            AbortReason
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

func AbortOtherActionsIfNil(reason AbortReason) Option {
	return func(action *basicAction) {
		action.abortOtherActionsIfNil = true
		action.abortReason = reason
	}
}

func (b *basicAction) Skip() {
	b.isFinished = true
}

func (b *basicAction) IsFinished() bool {
	if b == nil {
		return true
	}

	return b.isFinished
}
