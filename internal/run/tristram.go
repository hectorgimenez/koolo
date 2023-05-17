package run

import (
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

func (a Tristram) BuildActions() (actions []action.Action) {
	// Moving to starting point (Stony Field)
	actions = append(actions, a.builder.WayPoint(area.StonyField))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to Tristram portal
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		for _, o := range d.Objects {
			if o.Name == object.CairnStoneAlpha {
				return []step.Step{
					step.MoveTo(o.Position),
					step.SyncStep(func(d data.Data) error {
						helper.Sleep(1000)
						return nil
					}),
				}
			}
		}

		return nil
	}))

	// Clear monsters around the portal
	if config.Config.Game.Tristram.ClearPortal {
		actions = append(actions, a.builder.ClearAreaAroundPlayer(10))
	}

	actions = append(actions, a.openPortalIfNotOpened())

	// Enter Tristram portal
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.InteractObject(object.PermanentTownPortal, func(d data.Data) bool {
				return d.PlayerUnit.Area == area.Tristram
			}),
			step.SyncStep(func(d data.Data) error {
				// Add small delay to fetch the monsters
				helper.Sleep(2000)
				return nil
			}),
		}
	}))

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear Tristram
	actions = append(actions, a.builder.ClearArea(false, data.MonsterAnyFilter()))

	return
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
				actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
					stone, _ := d.Objects.FindOne(st)
					return []step.Step{
						step.InteractObject(stone.Name, nil),
						step.SyncStep(func(d data.Data) error {
							helper.Sleep(500)
							return nil
						}),
					}
				}))
			}
		}

		// Wait until portal is open
		actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
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
