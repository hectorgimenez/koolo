package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
)

type StonyTomb struct {
	baseRun
}

func (a StonyTomb) Name() string {
	return "StonyTomb"
}

func (a StonyTomb) BuildActions() (actions []action.Action) {
	actions = append(actions,
		a.builder.WayPoint(area.DryHills),
		a.builder.MoveToArea(area.RockyWaste),
		a.builder.MoveToArea(area.StonyTombLevel1),
	)

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.NewStepChain(func(_ data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear both level of Stony Tomb
	actions = append(actions,
		a.builder.ClearArea(true, data.MonsterAnyFilter()),
		a.builder.MoveToArea(area.StonyTombLevel2),
		a.builder.ClearArea(true, data.MonsterAnyFilter()),
	)

	return
}
