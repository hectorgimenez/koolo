package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	fixedPlaceNearRedPortalX = 5130
	fixedPlaceNearRedPortalY = 5120

	safeDistanceFromPindleX = 10056
	safeDistanceFromPindleY = 13239
)

type Pindleskin struct {
	BaseRun
}

func NewPindleskin(run BaseRun) Pindleskin {
	return Pindleskin{
		BaseRun: run,
	}
}

func (p Pindleskin) Name() string {
	return "Pindleskin"
}

func (p Pindleskin) BuildActions(data game.Data) (actions []action.Action) {
	if data.Area != game.AreaHarrogath {
		actions = append(actions, p.builder.WayPoint(game.AreaHarrogath))
	}

	actions = append(actions, action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		steps = append(steps,
			step.NewMoveTo(fixedPlaceNearRedPortalX, fixedPlaceNearRedPortalY, false, p.BaseRun.pf),
			step.NewInteractObject("PermanentTownPortal", func(data game.Data) bool {
				return data.Area == game.AreaNihlathaksTemple
			}, p.BaseRun.pf),
		)
		return
	}))

	return
}

func (p Pindleskin) Kill() error {
	err := p.char.KillPindle()
	if err != nil {
		return err
	}

	return nil
}

func (p Pindleskin) MoveToStartingPoint() error {
	//if game.Status().Area != game.AreaHarrogath {
	//	if err := p.tm.WPTo(5, 1); err != nil {
	//		return err
	//	}
	//}
	//
	//portal, found := p.getRedPortal()
	//if !found {
	//	// Let's do a first approach via static pathing, looks like portal is too far away
	//	p.pf.MoveTo(fixedPlaceNearRedPortalX, fixedPlaceNearRedPortalY, false)
	//
	//	portal, found = p.getRedPortal()
	//	if !found {
	//		return errors.New("portal not found")
	//	}
	//}
	//
	//err := p.pf.InteractToObject(portal, func(data game.Data) bool {
	//	time.Sleep(time.Second)
	//	return game.Status().Area == game.AreaNihlathaksTemple
	//})
	//if err != nil {
	//	return errors.New("error moving to red portal")
	//}
	//
	//p.char.Buff()
	return nil
}

func (p Pindleskin) TravelToDestination() error {
	//p.pf.MoveTo(safeDistanceFromPindleX, safeDistanceFromPindleY, true)
	//
	return nil
}

func (p Pindleskin) getRedPortal() (game.Object, bool) {
	for _, o := range game.Status().Objects {
		if o.IsRedPortal() {
			return o, true
		}
	}

	return game.Object{}, false
}
