package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) Heal() *StaticAction {
	return BuildStatic(func(data game.Data) (steps []step.Step) {
		shouldHeal := false
		if data.PlayerUnit.HPPercent() < 80 {
			b.logger.Info(fmt.Sprintf("Current life is %d%%, healing on NPC", data.PlayerUnit.HPPercent()))
			shouldHeal = true
		}

		if data.PlayerUnit.HasDebuff() {
			b.logger.Info(fmt.Sprintf("Debuff detected, healing on NPC"))
			shouldHeal = true
		}

		if shouldHeal {
			steps = append(steps, step.InteractNPC(town.GetTownByArea(data.PlayerUnit.Area).HealNPC()))
		}

		return
	}, CanBeSkipped())
}
