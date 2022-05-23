package action

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/hid"
	"time"
)

const (
	wpTabStartX     = 21.719
	wpTabStartY     = 7.928
	wpListStartX    = 5.37
	wpListStartY    = 6.852
	wpTabSize       = 69
	wpAreaBtnHeight = 49
)

func (b Builder) WayPoint(area game.Area) *StaticAction {
	allowedAreas := map[game.Area][2]int{
		game.AreaBlackMarsh:          {1, 5},
		game.AreaCatacombsLevel2:     {1, 9},
		game.AreaLostCity:            {2, 6},
		game.AreaArcaneSanctuary:     {2, 8},
		game.AreaDuranceOfHateLevel2: {3, 9},
		game.AreaHarrogath:           {5, 1},
		game.AreaHallsOfPain:         {5, 6},
		game.AreaTravincal:           {3, 8},
	}

	return BuildStatic(func(data game.Data) (steps []step.Step) {
		// We don't need to move
		if data.Area == area {
			return
		}

		wpCoords, found := allowedAreas[area]
		if !found {
			panic("Area destination is not mapped on WayPoint Action (waypoint.go)")
		}

		for _, o := range data.Objects {
			if o.IsWaypoint() {
				steps = append(steps,
					step.InteractObject(o.Name, func(data game.Data) bool {
						return data.OpenMenus.Waypoint
					}),
					step.SyncStep(func(data game.Data) error {
						actTabX := int(float32(hid.GameAreaSizeX)/wpTabStartX) + (wpCoords[0]-1)*wpTabSize + (wpTabSize / 2)
						actTabY := int(float32(hid.GameAreaSizeY) / wpTabStartY)

						areaBtnX := int(float32(hid.GameAreaSizeX) / wpListStartX)
						areaBtnY := int(float32(hid.GameAreaSizeY)/wpListStartY) + (wpCoords[1]-1)*wpAreaBtnHeight + (wpAreaBtnHeight / 2)
						hid.MovePointer(actTabX, actTabY)
						time.Sleep(time.Millisecond * 200)
						hid.Click(hid.LeftButton)
						time.Sleep(time.Millisecond * 200)
						hid.MovePointer(areaBtnX, areaBtnY)
						time.Sleep(time.Millisecond * 200)
						hid.Click(hid.LeftButton)

						for i := 0; i < 10; i++ {
							d, _ := game.Status()
							if d.Area == area {
								// Give some time to load the area
								time.Sleep(time.Second * 4)
								return nil
							}
							time.Sleep(time.Second * 1)
						}

						return errors.New("error changing area zone")
					}),
				)
			}
		}

		return
	}, Resettable())
}
