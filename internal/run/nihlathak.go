package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/pather"
	"go.uber.org/zap"
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

	// Move close to Nilhatak, but don't teleport over all the monsters
	var nihlathakStartPosition game.Object
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		for _, o := range data.Objects {
			if o.Name == object.NihlathakWildernessStartPositionName {
				nihlathakStartPosition = o
				return []step.Step{step.MoveTo(o.Position.X, o.Position.Y, true, step.StopAtDistance(25))}
			}
		}

		return []step.Step{}
	}))

	// Kill Nihlathak
	actions = append(actions, a.char.KillNihlathak())

	// Clear monsters around the area, sometimes it makes difficult to pickup items if there are many monsters around the area
	actions = append(actions, action.BuildDynamic(func(data game.Data) ([]step.Step, bool) {
		if config.Config.Game.Nihlathak.ClearArea {
			for _, m := range data.Monsters.Enemies() {
				if d := pather.DistanceFromPoint(nihlathakStartPosition.Position.X, nihlathakStartPosition.Position.Y, m.Position.X, m.Position.Y); d < 15 {
					a.logger.Debug("Clearing monsters around Nihlathak position", zap.Any("monster", m))
					return a.char.KillMonsterSequence(data, m.UnitID), true
				}
			}
		}

		return nil, false
	}))
	return
}
