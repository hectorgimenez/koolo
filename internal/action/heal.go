package action

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/town"
)

func (b Builder) Heal() *BasicAction {
	return BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if data.Health.HPPercent() < 80 {
			b.logger.Info(fmt.Sprintf("Current life is %d, healing on NPC", data.Health.HPPercent()))
			steps = append(steps, step.InteractNPC(town.GetTownByArea(data.Area).RefillNPC()))
		}

		return
	})
}
