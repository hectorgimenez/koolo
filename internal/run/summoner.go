package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Summoner struct {
	baseRun
}

func (s Summoner) Name() string {
	return "Summoner"
}

func (s Summoner) BuildActions(data game.Data) (actions []action.Action) {
	// Moving to starting point (Arcane Sanctuary)
	if data.Area != game.ArcaneSanctuary {
		actions = append(actions, s.builder.WayPoint(game.ArcaneSanctuary))
	}

	// Buff
	actions = append(actions, s.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		npc, found := game.Status().NPCs[game.TheSummoner]
		if !found {
			return nil
		}

		return []step.Step{
			step.MoveTo(npc.Positions[0].X, npc.Positions[0].Y, true),
		}
	}))

	// Kill Summoner
	actions = append(actions, s.char.KillSummoner())
	return
}
