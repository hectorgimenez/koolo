package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
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
		s.builder.MoveTo(func(d data.Data) (data.Position, bool) {
			for _, al := range d.AdjacentLevels {
				if al.Area == area.DuranceOfHateLevel1 {
					return data.Position{
						X: al.Position.X - 1,
						Y: al.Position.Y + 3,
					}, true
				}
			}
			return data.Position{}, false
		}),
		s.char.KillCouncil(), // Kill Council
	}
}
