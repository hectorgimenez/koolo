package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

var talRashaTombs = []area.Area{area.TalRashasTomb1, area.TalRashasTomb2, area.TalRashasTomb3, area.TalRashasTomb4, area.TalRashasTomb5, area.TalRashasTomb6, area.TalRashasTomb7}

type TalRashaTombs struct {
	baseRun
}

func (a TalRashaTombs) Name() string {
	return "TalRashaTombs"
}

func (a TalRashaTombs) BuildActions() (actions []action.Action) {
	a.logger.Info("Starting Tal Rasha Tombs run....")

	for _, tomb := range talRashaTombs {
		actions = append(actions,
			a.builder.WayPoint(area.CanyonOfTheMagi),
			a.builder.MoveToArea(tomb),
			a.builder.Buff(),
			a.builder.ClearArea(true, data.MonsterAnyFilter()),
			a.builder.ReturnTown(),
		)
		actions = append(actions, a.builder.PreRun(false)...)
	}

	return actions
}
