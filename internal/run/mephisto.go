package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

const (
	safeDistanceFromMephistoX = 17568
	safeDistanceFromMephistoY = 8069
)

type Mephisto struct {
	baseRun
}

func (m Mephisto) Name() string {
	return "Mephisto"
}

func (m Mephisto) BuildActions() (actions []action.Action) {
	// Moving to starting point (Durance of Hate Level 2)
	actions = append(actions, m.builder.WayPoint(area.DuranceOfHateLevel2))

	// Buff
	actions = append(actions, m.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.DuranceOfHateLevel3),
			step.MoveTo(safeDistanceFromMephistoX, safeDistanceFromMephistoY, true),
		}
	}))

	// Kill Mephisto
	actions = append(actions, m.char.KillMephisto())
	return
}
