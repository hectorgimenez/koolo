package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type Tristram struct {
	baseRun
}

func (a Tristram) Name() string {
	return "Tristram"
}

func (a Tristram) BuildActions() []action.Action {
	actions := []action.Action{
		a.builder.WayPoint(area.StonyField), // Moving to starting point (Stony Field)
		action.NewChain(func(d data.Data) []action.Action {
			for _, o := range d.Objects {
				if o.Name == object.CairnStoneAlpha {
					return []action.Action{a.builder.MoveToCoords(o.Position)}
				}
			}

			return nil
		}),
	}

	// Clear monsters around the portal
	if config.Config.Game.Tristram.ClearPortal {
		actions = append(actions, a.builder.ClearAreaAroundPlayer(10))
	}

	actions = append(actions, a.openPortalIfNotOpened())

	// Enter Tristram portal
	actions = append(actions, a.builder.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
		return d.PlayerUnit.Area == area.Tristram
	}, step.Wait(time.Second)))

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear Tristram or rescue Cain
	return append(actions, action.NewChain(func(d data.Data) []action.Action {
		if o, found := d.Objects.FindOne(object.CainGibbet); found && o.Selectable {
			return []action.Action{a.builder.InteractObject(object.CainGibbet, func(d data.Data) bool {
				obj, _ := d.Objects.FindOne(object.CainGibbet)

				return !obj.Selectable
			})}
		} else {
			filter := data.MonsterAnyFilter()
			if config.Config.Game.Tristram.FocusOnElitePacks {
				filter = data.MonsterEliteFilter()
			}
			return []action.Action{a.builder.ClearArea(false, filter)}
		}
	}))
}

func (a Tristram) openPortalIfNotOpened() action.Action {
	logged := false

	return action.NewChain(func(d data.Data) (actions []action.Action) {
		_, found := d.Objects.FindOne(object.PermanentTownPortal)
		if found {
			return nil
		}

		if !logged {
			a.logger.Debug("Tristram portal not detected, trying to open it")
			logged = true
		}

		// We don't know which order the stones are, so we activate all of them one by one in sequential order, 5 times
		for i := 0; i < 5; i++ {
			for _, cainStone := range []object.Name{
				object.CairnStoneAlpha,
				object.CairnStoneBeta,
				object.CairnStoneGamma,
				object.CairnStoneDelta,
				object.CairnStoneLambda,
			} {
				st := cainStone
				stone, _ := d.Objects.FindOne(st)
				actions = append(actions, a.builder.InteractObject(stone.Name, nil))
			}
		}

		// Wait until portal is open
		actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
			return []step.Step{
				step.SyncStepWithCheck(func(d data.Data) error {
					helper.Sleep(1000)
					return nil
				}, func(d data.Data) step.Status {
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
