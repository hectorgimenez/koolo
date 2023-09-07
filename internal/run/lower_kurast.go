package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

type LowerKurast struct {
	baseRun
}

func (a LowerKurast) Name() string {
	return "LowerKurast"
}

func (a LowerKurast) BuildActions() (actions []action.Action) {
	return []action.Action{
		a.builder.WayPoint(area.LowerKurast), // Moving to starting point (Lower Kurast)
		a.builder.ClearArea(true, data.MonsterEliteFilter()),
	}
}
