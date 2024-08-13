package run

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Tristram struct {
	baseRun
}

func (a Tristram) Name() string {
	return string(config.TristramRun)
}

func (a Tristram) BuildActions() []action.Action {
	actions := []action.Action{
		a.builder.WayPoint(area.StonyField), // Moving to starting point (Stony Field)
		action.NewChain(func(d game.Data) []action.Action {
			for _, o := range d.Objects {
				if o.Name == object.CairnStoneAlpha {
					return []action.Action{a.builder.MoveToCoords(o.Position)}
				}
			}

			return nil
		}),
	}

	// Clear monsters around the portal
	if a.CharacterCfg.Game.Tristram.ClearPortal || a.CharacterCfg.Game.Runs[0] == "leveling" {
		actions = append(actions, a.builder.ClearAreaAroundPlayer(20, data.MonsterAnyFilter()))
	}

	actions = append(actions, a.openPortalIfNotOpened())

	// Enter Tristram portal
	actions = append(actions, a.builder.InteractObject(object.PermanentTownPortal, func(d game.Data) bool {
		return d.PlayerUnit.Area == area.Tristram
	}, step.Wait(time.Second)))

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	// Clear Tristram or rescue Cain
	return append(actions, action.NewChain(func(d game.Data) []action.Action {
		if o, found := d.Objects.FindOne(object.CainGibbet); found && o.Selectable {
			return []action.Action{a.builder.InteractObject(object.CainGibbet, func(d game.Data) bool {
				obj, _ := d.Objects.FindOne(object.CainGibbet)

				return !obj.Selectable
			})}
		} else {
			filter := data.MonsterAnyFilter()
			if a.CharacterCfg.Game.Tristram.FocusOnElitePacks && a.CharacterCfg.Game.Runs[0] != "leveling" {
				filter = data.MonsterEliteFilter()
			}
			return []action.Action{a.builder.ClearArea(false, filter)}
		}
	}))
}

func (a Tristram) openPortalIfNotOpened() action.Action {
	logged := false

	return action.NewChain(func(d game.Data) (actions []action.Action) {
		_, found := d.Objects.FindOne(object.PermanentTownPortal)
		if found {
			return nil
		}

		if !logged {
			a.logger.Debug("Tristram portal not detected, trying to open it")
			logged = true
		}

		// We don't know which order the stones are, so we activate all of them one by one in sequential order, 5 times
		//activeStones := 0
		for range 6 {
			stoneTries := 0
			activeStones := 0
			for _, cainStone := range []object.Name{
				object.CairnStoneAlpha,
				object.CairnStoneBeta,
				object.CairnStoneGamma,
				object.CairnStoneDelta,
				object.CairnStoneLambda,
			} {
				st := cainStone
				stone, _ := d.Objects.FindOne(st)
				if stone.Selectable {
					actions = append(actions, a.builder.InteractObject(stone.Name, func(d game.Data) bool {
						if stoneTries < 5 {
							stoneTries++
							helper.Sleep(200)
							x, y := a.PathFinder.GameCoordsToScreenCords(d.PlayerUnit.Position.X, d.PlayerUnit.Position.Y, stone.Position.X, stone.Position.Y)
							a.HID.Click(game.LeftButton, x+3*stoneTries, y)
							a.logger.Debug(fmt.Sprintf("Tried to click %s at screen pos %vx%v", stone.Desc().Name, x, y))
							return false
						}
						stoneTries = 0
						return true
					}),
					)
				} else {
					helper.Sleep(500)
					activeStones++
				}
				_, tristPortal := d.Objects.FindOne(object.PermanentTownPortal)
				if activeStones >= 5 || tristPortal {
					break
				}
			}

		}

		// Sometimes the portal is out of detection range for some reason, this way it moves to the stones and enters the portal.
		st, stone := d.Objects.FindOne(object.CairnStoneAlpha)
		if stone {
			actions = append(actions, a.builder.MoveToCoords(st.Position))
			actions = append(actions, a.builder.InteractObject(object.PermanentTownPortal, func(d game.Data) bool {
				return d.PlayerUnit.Area == area.Tristram
			}, step.Wait(time.Second)))

			return actions
		}

		// Wait until portal is open
		actions = append(actions, action.NewStepChain(func(d game.Data) []step.Step {
			return []step.Step{
				step.SyncStepWithCheck(func(d game.Data) error {
					helper.Sleep(1000)
					return nil
				}, func(d game.Data) step.Status {
					_, found := d.Objects.FindOne(object.PermanentTownPortal)
					if !found {
						return step.StatusInProgress
					}
					return step.StatusCompleted
				}),
			}
		}))

		return actions
	})
}
