package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Action interface {
	Finished(data game.Data) bool
	NextStep(data game.Data) error
}

type BasicAction struct {
	Steps           []step.Step
	builder         func(data game.Data) []step.Step
	builderExecuted bool
}

func BuildOnRuntime(builder func(data game.Data) []step.Step) *BasicAction {
	return &BasicAction{
		builder: builder,
	}
}

func (b *BasicAction) Finished(data game.Data) bool {
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
			return s.Run(data)
		}
	}

	return nil
}
