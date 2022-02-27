package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
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
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		for _, o := range data.Objects {
			if o.Name == "CompellingOrb" {
				steps = append(steps,
					step.MoveTo(o.Position.X, o.Position.Y+10, true),
				)
			}
		}

		return
	}))

	// Kill Council
	actions = append(actions, s.char.KillCouncil())
	return
}
