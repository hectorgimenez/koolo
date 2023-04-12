package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
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
	actions = append(actions, action.BuildStatic(func(d data.Data) (steps []step.Step) {
		for _, al := range d.AdjacentLevels {
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
