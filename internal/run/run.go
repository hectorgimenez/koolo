package run

import (
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/game/data"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Run interface {
	MoveToStartingPoint() error
	TravelToDestination() error
	Kill() error
}

type BaseRun struct {
	dr   data.DataRepository
	pf   helper.PathFinder
	char character.Character
}

func NewBaseRun(dr data.DataRepository, pf helper.PathFinder, char character.Character) BaseRun {
	return BaseRun{
		dr:   dr,
		pf:   pf,
		char: char,
	}
}

func (br BaseRun) UseTP() {
	panic("implement TP")
	// TODO: Implement TP usage
}
