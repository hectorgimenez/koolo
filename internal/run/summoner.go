package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

type Summoner struct {
	baseRun
}

func (s Summoner) Name() string {
	return "Summoner"
}

func (s Summoner) BuildActions() (actions []action.Action) {
	// Moving to starting point (Arcane Sanctuary)
	actions = append(actions, s.builder.WayPoint(area.ArcaneSanctuary))

	// Buff
	actions = append(actions, s.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		m, found := d.NPCs.FindOne(npc.Summoner)
		if !found {
			return nil
		}

		return []step.Step{
			step.MoveTo(m.Positions[0]),
		}
	}, action.CanBeSkipped()))

	// Kill Summoner
	actions = append(actions, s.char.KillSummoner())
	return
}
