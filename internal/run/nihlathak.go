package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/object"
)

type Nihlathak struct {
	baseRun
}

func (a Nihlathak) Name() string {
	return "Nihlathak"
}

func (a Nihlathak) BuildActions() (actions []action.Action) {
	// Moving to starting point (Halls of Pain)
	actions = append(actions, a.builder.WayPoint(area.HallsOfPain))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to boss position
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.HallsOfVaught),
		}
	}))

	// Move to Nilhatak
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		for _, o := range data.Objects {
			if o.Name == object.NihlathakWildernessStartPositionName {
				return []step.Step{step.MoveTo(o.Position.X, o.Position.Y, true)}
			}
		}

		return []step.Step{}
	}))

	// Kill Nihlathak
	actions = append(actions, a.char.KillNihlathak())
	return
}
