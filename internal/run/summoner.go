package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/npc"
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
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		m, found := data.NPCs.FindOne(npc.Summoner)
		if !found {
			return nil
		}

		return []step.Step{
			step.MoveTo(m.Positions[0].X, m.Positions[0].Y, true),
		}
	}, action.CanBeSkipped()))

	// Kill Summoner
	actions = append(actions, s.char.KillSummoner())
	return
}
