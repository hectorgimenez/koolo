package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type StonyTomb struct {
	baseRun
}

func (a StonyTomb) Name() string {
	return "StonyTomb"
}

func (a StonyTomb) BuildActions() (actions []action.Action) {
	// actions = append(actions, a.builder.MoveToAreaAndKill(area.))
	actions = append(actions,
		a.builder.WayPoint(area.DryHills),
		a.char.Buff(),
		a.builder.MoveToArea(area.RockyWaste),
		a.builder.MoveToArea(area.StonyTombLevel1),
	)

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear StonyTombLevel1
	actions = append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))

	// Travel to StonyTombLevel2
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.StonyTombLevel2),
			step.SyncStep(func(_ data.Data) error {
				// Add small delay to fetch the monsters
				helper.Sleep(500)
				return nil
			}),
		}
	}))

	// Clear StonyTombLevel2
	actions = append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))

	return
}
