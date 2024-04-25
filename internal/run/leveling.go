package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/health"
)

type Leveling struct {
	baseRun
	bm health.BeltManager
}

func (a Leveling) Name() string {
	return string(config.LevelingRun)
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
