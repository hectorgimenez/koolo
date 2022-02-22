package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	fixedPositionInsideCouncilBuildingX = 4817
	fixedPositionInsideCouncilBuildingY = 2444
)

type Council struct {
	baseRun
}

func (s Council) Name() string {
	return "Council"
}

func (s Council) BuildActions() (actions []action.Action) {
	// Moving to starting point (Travincal)
	actions = append(actions, s.builder.WayPoint(game.AreaTravincal))

	// Buff
	actions = append(actions, s.char.Buff())

	// Travel to boss position
	// Travel to boss position
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveTo(fixedPositionInsideCouncilBuildingX, fixedPositionInsideCouncilBuildingY, true),
		}
	}))

	// Kill Council
	actions = append(actions, s.char.KillCouncil())
	return
}
