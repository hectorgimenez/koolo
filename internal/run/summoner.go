package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Summoner struct {
	BaseRun
}

func NewSummoner(run BaseRun) Summoner {
	return Summoner{
		BaseRun: run,
	}
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
			step.MoveTo(npc.Positions[0].X, npc.Positions[1].X, true),
		}
	}))

	// Kill Summoner
	actions = append(actions, s.char.KillSummoner())
	return
}

//func (p Summoner) Kill() error {
//	err := p.char.KillSummoner()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (p Summoner) MoveToStartingPoint() error {
//	if game.Status().Area != game.ArcaneSanctuary {
//		if err := p.tm.WPTo(2, 8); err != nil {
//			return err
//		}
//	}
//
//	if game.Status().Area != game.ArcaneSanctuary {
//		return errors.New("error moving to Arcane Sanctuary")
//	}
//
//	p.char.Buff()
//	return nil
//}
//
//func (p Summoner) TravelToDestination() error {
//	npc, found := game.Status().NPCs[game.TheSummoner]
//	if !found {
//		return errors.New("The Summoner not found")
//	}
//
//	p.pf.MoveTo(npc.Positions[0].X, npc.Positions[0].Y, true)
//
//	return nil
//}
