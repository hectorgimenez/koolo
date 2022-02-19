package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/character"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Run interface {
	Name() string
	BuildActions(game.Data) []action.Action
}

type BaseRun struct {
	builder action.Builder
	char    character.Character
}

func NewBaseRun(builder action.Builder, char character.Character) BaseRun {
	return BaseRun{
		builder: builder,
		char:    char,
	}
}
