package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
)

type Council struct {
	baseRun
}

func (s Council) Name() string {
	return "Council"
}

func (s Council) BuildActions() (actions []action.Action) {
	// Moving to starting point (Travincal)
	actions = append(actions, s.builder.WayPoint(area.Travincal))

	// Buff
	actions = append(actions, s.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(data game.Data) (steps []step.Step) {
		for _, al := range data.AdjacentLevels {
			if al.Area == area.DuranceOfHateLevel1 {
				steps = append(steps, step.MoveTo(al.Position.X-1, al.Position.Y+3, true))
			}
		}

		return
	}))

	// Kill Council
	actions = append(actions, s.char.KillCouncil())
	return
}
