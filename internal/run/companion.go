package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/town"
)

type Companion struct {
	baseRun
}

func (s Companion) Name() string {
	return "Companion"
}

func (s Companion) BuildActions() (actions []action.Action) {
	actions = append(actions, s.builder.MoveTo(func(d data.Data) (data.Position, bool) {
		tpArea := town.GetTownByArea(d.PlayerUnit.Area).TPWaitingArea(d)

		return tpArea, true
	}))

	// Wait for the portal, once it's up, enter and wait
	portalFound := false
	actions = append(actions, action.NewStepChain(func(d data.Data) []step.Step {
		if !portalFound {
			for _, o := range d.Objects {
				if o.IsPortal() {
					portalFound = true
					return []step.Step{
						step.MoveTo(o.Position),
						step.InteractObject(o.Name, func(d data.Data) bool {
							return !d.PlayerUnit.Area.IsTown()
						}),
					}
				}
			}

			return []step.Step{step.Wait(time.Second)}
		}

		return nil
	}))

	actions = append(actions, action.NewChain(func(d data.Data) []action.Action {
		rm, found := d.Roster.FindByName(config.Config.Companion.LeaderName)
		if !found {
			return nil
		}

		if d.PlayerUnit.Area == area.ThroneOfDestruction {
			return []action.Action{s.builder.Wait(time.Second)}
		}

		return []action.Action{s.builder.MoveToCoords(rm.Position)}
	}))

	return actions
}
