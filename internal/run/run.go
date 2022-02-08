package run

import (
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/helper"
	"log"
)

type Run interface {
	MoveToStartingPoint() error
	TravelToDestination() error
	Kill() error
	ReturnToTown()
}

type BaseRun struct {
	dr   data.DataRepository
	pf   helper.PathFinder
	char character.Character
}

func (b BaseRun) ReturnToTown() {
	b.char.UseTP()
	log.Println("Entering Portal...")
	for _, o := range b.dr.GameData().Objects {
		if o.IsPortal() {
			b.pf.InteractToObject(o)
		}
	}
}

func NewBaseRun(dr data.DataRepository, pf helper.PathFinder, char character.Character) BaseRun {
	return BaseRun{
		dr:   dr,
		pf:   pf,
		char: char,
	}
}
