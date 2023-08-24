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

func (s Council) BuildActions() []action.Action {
	return []action.Action{
		s.builder.WayPoint(area.Travincal), // Moving to starting point (Travincal)
		s.char.Buff(),                      // Buff
		action.BuildStatic(func(d data.Data) (steps []step.Step) {
			for _, al := range d.AdjacentLevels {
				if al.Area == area.DuranceOfHateLevel1 {
					steps = append(steps, step.MoveTo(data.Position{
						X: al.Position.X - 1,
						Y: al.Position.Y + 3,
					}))
				}
			}

			return
		}),
		s.char.KillCouncil(), // Kill Council
	}
}
