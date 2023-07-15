package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

var mephistoSafePosition = data.Position{
	X: 17568,
	Y: 8069,
}

type Mephisto struct {
	baseRun
}

func (m Mephisto) Name() string {
	return "Mephisto"
}

func (m Mephisto) BuildActions() []action.Action {
	return []action.Action{
		// Moving to starting point (Durance of Hate Level 2)
		m.builder.WayPoint(area.DuranceOfHateLevel2),
		m.char.Buff(),
		// Travel to boss position
		m.builder.MoveToArea(area.DuranceOfHateLevel3),
		m.builder.MoveToCoords(mephistoSafePosition),
		// Kill Mephisto
		m.char.KillMephisto(),
	}
}
