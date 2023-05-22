package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

func (a Leveling) act2() (actions []action.Action) {
	actions = append(actions,
		//a.radament(),
		a.findHoradricCube(),
	)

	return
}

func (a Leveling) radament() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		actions = append(actions,
			a.builder.WayPoint(area.SewersLevel2Act2),
			a.builder.MoveToArea(area.SewersLevel3Act2),
		)

		// TODO: Find Radament (use 355 object to locate him)
		return
	})
}

func (a Leveling) findHoradricCube() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		for _, i := range d.Items.Inventory {
			if i.Name == "HoradricCube" {
				a.logger.Info("Horadric Cube found, skipping quest")
				return nil
			}
		}

		a.logger.Info("Horadric Cube not found, starting quest")
		return []action.Action{
			a.builder.MoveToArea(area.RockyWaste),
			a.builder.MoveToArea(area.DryHills),
			a.builder.DiscoverWaypoint(),
			a.builder.MoveToArea(area.HallsOfTheDeadLevel1),
			a.builder.MoveToArea(area.HallsOfTheDeadLevel2),
			a.builder.DiscoverWaypoint(),
			a.builder.MoveToArea(area.HallsOfTheDeadLevel3),
			a.builder.ClearArea(true, data.MonsterAnyFilter()),
		}
	})
}
