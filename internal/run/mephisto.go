package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	safeDistanceFromMephistoX = 17568
	safeDistanceFromMephistoY = 8069
)

type Mephisto struct {
	BaseRun
}

func NewMephisto(run BaseRun) Mephisto {
	return Mephisto{
		BaseRun: run,
	}
}

func (m Mephisto) Name() string {
	return "Mephisto"
}

func (m Mephisto) BuildActions(data game.Data) (actions []action.Action) {
	// Moving to starting point (Catacombs Level 2)
	if data.Area != game.AreaDuranceOfHateLevel2 {
		actions = append(actions, m.builder.WayPoint(game.AreaDuranceOfHateLevel2))
	}

	// Buff
	actions = append(actions, m.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(game.AreaDuranceOfHateLevel3),
			step.MoveTo(safeDistanceFromMephistoX, safeDistanceFromMephistoY, true),
		}
	}))

	// Kill Mephisto
	actions = append(actions, m.char.KillMephisto())
	return
}
