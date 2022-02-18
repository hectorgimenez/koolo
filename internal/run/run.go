package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Run interface {
	Name() string
	BuildActions(game.Data) []action.Action
}

type BaseRun struct {
	builder action.Builder
	pf      helper.PathFinderV2
	char    character.Character
}

func NewBaseRun(builder action.Builder, pf helper.PathFinderV2, char character.Character) BaseRun {
	return BaseRun{
		builder: builder,
		pf:      pf,
		char:    char,
	}
}
