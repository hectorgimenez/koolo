package action

import (
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) SwitchToLegacyMode() *StepChainAction {
	b.Logger.Debug("Switching to legacy mode...")
	return NewStepChain(func(d game.Data) []step.Step {
		return []step.Step{step.KeySequence(b.Reader.GetKeyBindings().LegacyToggle.Key1[0])}
	})
}
