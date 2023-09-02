package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/object"
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
	actions = append(actions, action.NewFactory(func(d data.Data) action.Action {
		if !portalFound {
			for _, o := range d.Objects {
				if o.IsPortal() {
					portalFound = true
				}
			}

			if portalFound {
				return s.builder.InteractObject(object.TownPortal, func(d data.Data) bool {
					return !d.PlayerUnit.Area.IsTown()
				})
			}

			return action.BuildStatic(func(d data.Data) []step.Step {
				return []step.Step{step.Wait(time.Second)}
			})
		}

		return nil
	}))

	actions = append(actions, action.NewFactory(func(d data.Data) action.Action {
		rm, found := d.Roster.FindByName(config.Config.Companion.LeaderName)
		if !found {
			return nil
		}

		if d.PlayerUnit.Area == area.ThroneOfDestruction {
			return s.builder.Wait(time.Second)
		}

		return s.builder.MoveToCoords(rm.Position)
	}))

	return actions
}
