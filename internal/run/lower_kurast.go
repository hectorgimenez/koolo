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
	// Moving to starting point (Lower Kurast)
	actions = append(actions, a.builder.WayPoint(area.LowerKurast))

	// Buff
	actions = append(actions, a.char.Buff())

	// Clear Lower Kurast
	actions = append(actions, a.builder.ClearArea(true, data.MonsterEliteFilter()))

	return
}
