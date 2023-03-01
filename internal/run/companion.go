package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/object"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/town"
	"time"
)

type Companion struct {
	baseRun
}

func (s Companion) Name() string {
	return "Companion"
}

func (s Companion) BuildActions() (actions []action.Action) {
	actions = append(actions, action.BuildStatic(func(data game.Data) []step.Step {
		tpArea := town.GetTownByArea(data.PlayerUnit.Area).TPWaitingArea(data)
		return []step.Step{step.MoveTo(tpArea.X, tpArea.Y, false)}
	}))

	// Wait for the portal, once it's up, enter and wait
	portalFound := false
	actions = append(actions, action.NewFactory(func(data game.Data) action.Action {
		if !portalFound {
			for _, o := range data.Objects {
				if o.IsPortal() {
					portalFound = true
				}
			}

			return action.BuildStatic(func(data game.Data) []step.Step {
				if portalFound {
					return []step.Step{
						step.InteractObject(object.TownPortal, func(data game.Data) bool {
							return !data.PlayerUnit.Area.IsTown()
						}),
					}
				}

				return []step.Step{step.SyncStep(func(data game.Data) error {
					return nil
				})}
			})
		}

		return nil
	}))

	actions = append(actions, action.NewFactory(func(data game.Data) action.Action {
		rm, found := data.Roster.FindByName(config.Config.Companion.LeaderName)
		if !found {
			return nil
		}

		return action.BuildStatic(func(data game.Data) []step.Step {
			if data.PlayerUnit.Area == area.ThroneOfDestruction {
				return []step.Step{step.SyncStep(func(data game.Data) error {
					helper.Sleep(1000)
					return nil
				})}
			} else {
				// Follow the leader
				return []step.Step{step.MoveTo(rm.Position.X, rm.Position.Y, false, step.WithTimeout(time.Second*3))}
			}
		})
	}))

	return actions
}
