package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type AncientTunnels struct {
	baseRun
}

func (a AncientTunnels) Name() string {
	return "AncientTunnels"
}

func (a AncientTunnels) BuildActions() (actions []action.Action) {
	// Moving to starting point (Lost City)
	actions = append(actions, a.builder.WayPoint(area.LostCity))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to ancient tunnels
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.AncientTunnels),
			step.SyncStep(func(_ data.Data) error {
				// Add small delay to fetch the monsters
				helper.Sleep(1000)
				return nil
			}),
		}
	}))

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear Ancient Tunnels
	actions = append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))

	return
}
