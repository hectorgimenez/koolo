package run

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"time"
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

func (p Pindleskin) Kill() error {
	err := p.char.KillPindle()
	if err != nil {
		p.UseTP()
		return err
	}

	return nil
}

func (p Pindleskin) MoveToStartingPoint() error {
	portal, found := p.getRedPortal()
	if !found {
		// Let's do a first approach via static pathing, looks like portal is too far away
		p.pf.MoveTo(fixedPlaceNearRedPortalX, fixedPlaceNearRedPortalY, false)

		portal, found = p.getRedPortal()
		if !found {
			return errors.New("portal not found")
		}
	}

	p.pf.InteractToObject(portal)
	time.Sleep(time.Second * 2)
	if p.dr.GameData().Area != data.AreaNihlathaksTemple {
		return errors.New("error moving to red portal")
	}

	p.char.Buff()
	return nil
}

func (p Pindleskin) TravelToDestination() error {
	p.pf.MoveTo(safeDistanceFromPindleX, safeDistanceFromPindleY, true)

	return nil
}

func (p Pindleskin) getRedPortal() (data.Object, bool) {
	d := p.dr.GameData()
	for _, o := range d.Objects {
		if o.IsRedPortal() {
			return o, true
		}
	}

	return data.Object{}, false
}
