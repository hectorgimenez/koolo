package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/reader"
)

type Leveling struct {
	baseRun
	gr *reader.GameReader
	bm health.BeltManager
}

func (a Leveling) Name() string {
	return "Leveling"
}

func (a Leveling) BuildActions() []action.Action {
	return []action.Action{
		a.act1(),
		a.act2(),
		a.act3(),
		a.act4(),
		a.act5(),
	}
}
