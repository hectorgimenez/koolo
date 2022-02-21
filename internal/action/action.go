package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const maxRetries = 5

type Action interface {
	Finished(data game.Data) bool
	NextStep(data game.Data) error
}

type BasicAction struct {
	Steps           []step.Step
	builder         func(data game.Data) []step.Step
	builderExecuted bool
	retries         int
}

func BuildOnRuntime(builder func(data game.Data) []step.Step) *BasicAction {
	return &BasicAction{
		builder: builder,
	}
}

func (b *BasicAction) Finished(data game.Data) bool {
	if b.retries >= maxRetries {
		return true
	}

	if b.builder != nil && !b.builderExecuted {
		return false
	}

	for _, s := range b.Steps {
		if s.Status(data) != step.StatusCompleted {
			return false
		}
	}

	return true
}

func (b *BasicAction) NextStep(data game.Data) error {
	if b.builder != nil && !b.builderExecuted {
		b.Steps = b.builder(data)
		b.builderExecuted = true
	}

	for _, s := range b.Steps {
		if s.Status(data) != step.StatusCompleted {
			err := s.Run(data)
			if err != nil {
				b.retries++
			}

			return err
		}
	}

	return nil
}
