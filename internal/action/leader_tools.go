package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

func (b *Builder) OpenTPIfLeader() *StepChainAction {
	isLeader := b.CharacterCfg.Companion.Enabled && b.CharacterCfg.Companion.Leader

	return NewStepChain(func(d data.Data) []step.Step {
		if isLeader {
			return []step.Step{step.OpenPortal(b.CharacterCfg.Bindings.TP)}
		}

		return []step.Step{step.Wait(50)}
	})
}
