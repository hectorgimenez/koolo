package run

import (
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
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
	tm   town.Manager
}

func (b BaseRun) ReturnToTown() {
	b.char.UseTP()
	for _, o := range game.Status().Objects {
		if o.IsPortal() {
			log.Println("Entering Portal...")
			b.pf.InteractToObject(o)
		}
	}
}

func NewBaseRun(pf helper.PathFinder, char character.Character, tm town.Manager) BaseRun {
	return BaseRun{
		pf:   pf,
		char: char,
		tm:   tm,
	}
}
