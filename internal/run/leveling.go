package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/reader"
)

type Leveling struct {
	baseRun
	gr *reader.GameReader
}

func (a Leveling) Name() string {
	return "Leveling"
}

func (a Leveling) BuildActions() []action.Action {
	return []action.Action{
		a.act1(),
		a.act2(),
	}
}
