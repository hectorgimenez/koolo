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
	pf   helper.PathFinder
	char character.Character
}

func (b BaseRun) ReturnToTown() {
	b.char.UseTP()
	for _, o := range data.Status.Objects {
		if o.IsPortal() {
			log.Println("Entering Portal...")
			b.pf.InteractToObject(o)
		}
	}
}

func NewBaseRun(pf helper.PathFinder, char character.Character) BaseRun {
	return BaseRun{
		pf:   pf,
		char: char,
	}
}
