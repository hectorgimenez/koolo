package action

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
)

const (
	wpTabStartX     = 130
	wpTabStartY     = 148
	wpTabSizeX      = 57
	wpListPositionX = 200
	wpListStartY    = 158
	wpAreaBtnHeight = 41
)

func (b Builder) WayPoint(a area.Area) *StaticAction {
	allowedAreas := map[area.Area][2]int{
		area.ColdPlains:              {1, 2},
		area.StonyField:              {1, 3},
		area.DarkWood:                {1, 4},
		area.BlackMarsh:              {1, 5},
		area.OuterCloister:           {1, 6},
		area.JailLevel1:              {1, 7},
		area.InnerCloister:           {1, 8},
		area.CatacombsLevel2:         {1, 9},
		area.LostCity:                {2, 6},
		area.ArcaneSanctuary:         {2, 8},
		area.LowerKurast:             {3, 5},
		area.Travincal:               {3, 8},
		area.DuranceOfHateLevel2:     {3, 9},
		area.RiverOfFlame:            {4, 3},
		area.Harrogath:               {5, 1},
		area.FrigidHighlands:         {5, 2},
		area.HallsOfPain:             {5, 6},
		area.TheWorldStoneKeepLevel2: {5, 9},
	}

	return BuildStatic(func(d data.Data) (steps []step.Step) {
		// We don't need to move
		if d.PlayerUnit.Area == a {
			return
		}

		wpCoords, found := allowedAreas[a]
		if !found {
			panic("Area destination is not mapped on WayPoint Action (waypoint.go)")
		}

		for _, o := range d.Objects {
			if o.IsWaypoint() {
				steps = append(steps,
					step.InteractObject(o.Name, func(d data.Data) bool {
						return d.OpenMenus.Waypoint
					}),
					step.SyncStepWithCheck(func(d data.Data) error {
						actTabX := wpTabStartX + (wpCoords[0]-1)*wpTabSizeX + (wpTabSizeX / 2)

						areaBtnY := wpListStartY + (wpCoords[1]-1)*wpAreaBtnHeight + (wpAreaBtnHeight / 2)
						hid.MovePointer(actTabX, wpTabStartY)
						helper.Sleep(200)
						hid.Click(hid.LeftButton)
						helper.Sleep(200)
						hid.MovePointer(wpListPositionX, areaBtnY)
						helper.Sleep(200)
						hid.Click(hid.LeftButton)

						return nil
					}, func(d data.Data) step.Status {
						if d.PlayerUnit.Area == a && !d.OpenMenus.LoadingScreen {
							return step.StatusCompleted
						}

						return step.StatusInProgress
					}),
				)
			}
		}

		return
	}, Resettable())
}
