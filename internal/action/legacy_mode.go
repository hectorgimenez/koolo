package action

import (
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (b *Builder) SwitchToLegacyMode() *StepChainAction {
	return NewStepChain(func(d game.Data) []step.Step {
		if d.CharacterCfg.ClassicMode && !d.LegacyGraphics {
			b.Logger.Debug("Switching to legacy mode...")
			return []step.Step{
				step.KeySequence(b.Reader.GetKeyBindings().LegacyToggle.Key1[0]),
				step.Wait(time.Millisecond * 500), // Add small delay to allow the game to switch
			}
		}
		return nil
	})
}
