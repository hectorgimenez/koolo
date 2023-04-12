package action

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) Heal() *StaticAction {
	return BuildStatic(func(d data.Data) (steps []step.Step) {
		shouldHeal := false
		if d.PlayerUnit.HPPercent() < 80 {
			b.logger.Info(fmt.Sprintf("Current life is %d%%, healing on NPC", d.PlayerUnit.HPPercent()))
			shouldHeal = true
		}

		if d.PlayerUnit.HasDebuff() {
			b.logger.Info(fmt.Sprintf("Debuff detected, healing on NPC"))
			shouldHeal = true
		}

		if shouldHeal {
			steps = append(steps, step.InteractNPC(town.GetTownByArea(d.PlayerUnit.Area).HealNPC()))
		}

		return
	}, CanBeSkipped())
}
