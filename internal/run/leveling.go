package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
)

type Leveling struct {
	baseRun
}

func (a Leveling) Name() string {
	return "Leveling"
}

func (a Leveling) BuildActions() (actions []action.Action) {
	actions = append(actions, a.act1())
	//actions = append(actions, a.act2()...)

	return
}
