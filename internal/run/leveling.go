package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/health"
)

const NameLeveling = "Leveling"

type Leveling struct {
	baseRun
	bm health.BeltManager
}

func (a Leveling) Name() string {
	return NameLeveling
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
