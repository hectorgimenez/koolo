package run

import (
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
)

type Run interface {
	MoveToStartingPoint() error
	TravelToDestination() error
	Kill() error
}

type BaseRun struct {
	pf   helper.PathFinder
	char character.Character
	tm   town.Manager
}

func NewBaseRun(pf helper.PathFinder, char character.Character, tm town.Manager) BaseRun {
	return BaseRun{
		pf:   pf,
		char: char,
		tm:   tm,
	}
}
