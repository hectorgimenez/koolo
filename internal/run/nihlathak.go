package run

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
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

	// Travel to boss position
	actions = append(actions, a.builder.MoveToArea(area.HallsOfVaught))

	// Move close to Nilhatak, but don't teleport over all the monsters
	var nilaO data.Object
	actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
		for _, o := range d.Objects {
			if o.Name == object.NihlathakWildernessStartPositionName {
				nilaO = o
				return []step.Step{step.MoveTo(o.Position, step.StopAtDistance(40))}
			}
		}

		return []step.Step{}
	}))

	// Try to find the safest place
	actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
		corners := [4]data.Position{
			{
				X: nilaO.Position.X + 13,
				Y: nilaO.Position.Y + 13,
			},
			{
				X: nilaO.Position.X - 13,
				Y: nilaO.Position.Y + 13,
			},
			{
				X: nilaO.Position.X - 13,
				Y: nilaO.Position.Y - 13,
			},
			{
				X: nilaO.Position.X + 13,
				Y: nilaO.Position.Y - 13,
			},
		}

		bestCorner := 0
		bestCornerDistance := 0
		for i, c := range corners {
			averageDistance := 0
			for _, m := range d.Monsters.Enemies() {
				averageDistance += pather.DistanceFromPoint(c, m.Position)
			}
			if averageDistance > bestCornerDistance {
				bestCorner = i
				bestCornerDistance = averageDistance
			}
			fmt.Printf("Corner %d. Average monster distance: %d\n", i, averageDistance)
		}

		fmt.Printf("Moving to corner %d. Average monster distance: %d\n", bestCorner, bestCornerDistance)
		return []step.Step{step.MoveTo(corners[bestCorner])}
	}))

	// Kill Nihlathak
	actions = append(actions, a.char.KillNihlathak())

	// Clear monsters around the area, sometimes it makes difficult to pickup items if there are many monsters around the area
	if config.Config.Game.Nihlathak.ClearArea {
		actions = append(actions, a.char.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
			for _, m := range d.Monsters.Enemies() {
				if d := pather.DistanceFromPoint(nilaO.Position, m.Position); d < 15 {
					a.logger.Debug("Clearing monsters around Nihlathak position", zap.Any("monster", m))
					return m.UnitID, true
				}
			}

			return 0, false
		}, nil))
	}

	return
}
