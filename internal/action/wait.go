package action

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

func (b *Builder) Wait(duration time.Duration) *StepChainAction {
	return NewStepChain(func(d data.Data) (steps []step.Step) {
		return []step.Step{
			step.Wait(duration),
		}
	})
}
