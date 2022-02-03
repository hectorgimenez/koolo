package run

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/helper"
	"time"
)

const (
	fixedPlaceNearRedPortalX = 5130
	fixedPlaceNearRedPortalY = 5120
)

type Pindleskin struct {
	dr data.DataRepository
	pf helper.PathFinder
}

func NewPindleskin(dr data.DataRepository, pf helper.PathFinder) Pindleskin {
	return Pindleskin{
		dr: dr,
		pf: pf,
	}
}

func (p Pindleskin) Kill() error {
	//TODO implement me
	panic("implement me")
}

func (p Pindleskin) MoveToStartingPoint() error {
	// Let's do a first approach to the portal before trying to detect it
	p.pf.MoveTo(fixedPlaceNearRedPortalX, fixedPlaceNearRedPortalY)
	portal, found := p.getRedPortal()
	if !found {
		return errors.New("portal not found")
	}

	p.pf.InteractToObject(portal)
	time.Sleep(time.Second * 3)
	if p.dr.GameData().Area == data.AreaNihlathaksTemple {
		return nil
	}

	return errors.New("error moving to red portal")
}

func (p Pindleskin) TravelToDestination() error {
	//TODO implement me
	panic("implement me")
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
