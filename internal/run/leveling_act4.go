package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

func (a Leveling) act4() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		if d.PlayerUnit.Area != area.ThePandemoniumFortress {
			return nil
		}

		quests := a.builder.GetCompletedQuests(4)
		if !quests[0] {
			return a.izual()
		}

		return Diablo{baseRun: a.baseRun, bm: a.bm}.BuildActions()
	})
}

func (a Leveling) izual() []action.Action {
	return []action.Action{
		a.builder.MoveToArea(area.OuterSteppes),
		a.char.Buff(),
		a.builder.MoveToArea(area.PlainsOfDespair),
		a.char.Buff(),
		a.builder.MoveTo(func(d data.Data) (data.Position, bool) {
			izual, found := d.NPCs.FindOne(npc.Izual)
			if !found {
				return data.Position{}, false
			}

			return izual.Positions[0], true
		}),
		a.char.KillIzual(),
		a.builder.ReturnTown(),
		action.BuildStatic(func(d data.Data) []step.Step {
			return []step.Step{
				step.InteractNPC(npc.Tyrael2),
			}
		}),
	}
}
