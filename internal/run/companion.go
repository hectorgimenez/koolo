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
	"github.com/hectorgimenez/koolo/internal/town"
)

type Companion struct {
	baseRun
}

func (s Companion) Name() string {
	return "Companion"
}

func (s Companion) BuildActions() (actions []action.Action) {
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		tpArea := town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)
		return []step.Step{step.MoveTo(tpArea)}
	}))

	// Wait for the portal, once it's up, enter and wait
	portalFound := false
	actions = append(actions, action.NewFactory(func(d data.Data) action.Action {
		if !portalFound {
			for _, o := range d.Objects {
				if o.IsPortal() {
					portalFound = true
				}
			}

			return action.BuildStatic(func(d data.Data) []step.Step {
				if portalFound {
					return []step.Step{
						step.InteractObject(object.TownPortal, func(d data.Data) bool {
							return !d.PlayerUnit.Area.IsTown()
						}),
					}
				}

				return []step.Step{step.SyncStep(func(d data.Data) error {
					return nil
				})}
			})
		}

		return nil
	}))

	actions = append(actions, action.NewFactory(func(d data.Data) action.Action {
		rm, found := d.Roster.FindByName(config.Config.Companion.LeaderName)
		if !found {
			return nil
		}

		return action.BuildStatic(func(d data.Data) []step.Step {
			if d.PlayerUnit.Area == area.ThroneOfDestruction {
				return []step.Step{step.SyncStep(func(d data.Data) error {
					helper.Sleep(1000)
					return nil
				})}
			} else {
				// Follow the leader
				return []step.Step{step.MoveTo(rm.Position, step.WithTimeout(time.Second*3))}
			}
		})
	}))

	return actions
}
