package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	andarielStartingPositionX = 22563
	andarielStartingPositionY = 9533
)

type Andariel struct {
	BaseRun
}

func NewAndariel(run BaseRun) Andariel {
	return Andariel{
		BaseRun: run,
	}
}

func (a Andariel) Name() string {
	return "Andariel"
}

func (a Andariel) BuildActions(data game.Data) (actions []action.Action) {
	// Moving to starting point (Catacombs Level 2)
	if data.Area != game.AreaCatacombsLevel2 {
		actions = append(actions, a.builder.WayPoint(game.AreaCatacombsLevel2))
	}

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildOnRuntime(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(game.AreaCatacombsLevel3),
			step.MoveToLevel(game.AreaCatacombsLevel4),
			step.MoveTo(andarielStartingPositionX, andarielStartingPositionY, true),
		}
	}))

	// Kill Andariel
	actions = append(actions, a.char.KillAndariel())
	return
}

//
//func (p Andariel) Kill() error {
//	err := p.char.KillAndariel()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (p Andariel) MoveToStartingPoint() error {
//	if err := p.tm.WPTo(1, 9); err != nil {
//		return err
//	}
//	time.Sleep(time.Second)
//
//	if game.Status().Area != game.AreaCatacombsLevel2 {
//		return errors.New("error moving to Catacombs Level 2")
//	}
//
//	p.char.Buff()
//	return nil
//}
//
//func (p Andariel) TravelToDestination() error {
//	err := p.pf.MoveToArea(game.AreaCatacombsLevel3)
//	if err != nil {
//		return err
//	}
//
//	err = p.pf.MoveToArea(game.AreaCatacombsLevel4)
//	if err != nil {
//		return err
//	}
//
//	p.pf.MoveTo(andarielStartingPositionX, andarielStartingPositionY, true)
//
//	return nil
//}
