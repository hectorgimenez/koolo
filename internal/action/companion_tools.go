package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) OpenTPIfLeader() *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		isLeader := d.CharacterCfg.Companion.Enabled && d.CharacterCfg.Companion.Leader
		if isLeader {
			return []step.Step{step.OpenPortal(d.CharacterCfg.Bindings.TP)}
		}

		return []step.Step{step.Wait(50)}
	})
}
