package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

var andarielStartingPosition = data.Position{
	X: 22561,
	Y: 9553,
}

type Andariel struct {
	baseRun
}

func (a Andariel) Name() string {
	return "Andariel"
}

func (a Andariel) BuildActions() []action.Action {
	return []action.Action{
		a.builder.WayPoint(area.CatacombsLevel2), // Moving to starting point (Catacombs Level 2)
		a.builder.MoveToArea(area.CatacombsLevel3),
		a.builder.MoveToArea(area.CatacombsLevel4),
		a.builder.MoveToCoords(andarielStartingPosition), // Travel to boss position
		a.char.KillAndariel(),                            // Kill Andariel
	}
}
