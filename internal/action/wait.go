package action

import (
	"github.com/hectorgimenez/koolo/internal/game"
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
)

func (b *Builder) Wait(duration time.Duration) *StepChainAction {
	return NewStepChain(func(d game.Data) (steps []step.Step) {
		return []step.Step{
			step.Wait(duration),
		}
	})
}
