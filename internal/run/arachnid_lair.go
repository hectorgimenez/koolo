package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
)

type ArachnidLair struct {
	baseRun
}

func (a ArachnidLair) Name() string {
	return "ArachnidLair"
}

func (a ArachnidLair) BuildActions() (actions []action.Action) {
	
	actions = append(actions, 
		// Moving to starting point (Spider FOrest)
		a.builder.WayPoint(area.SpiderForest),

		// Buff
		a.char.Buff(),

		// Travel to ArachnidLair
		a.builder.MoveToArea(area.SpiderCave),
	)

	if config.Config.Companion.Enabled && config.Config.Companion.Leader {
		actions = append(actions, action.BuildStatic(func(_ data.Data) []step.Step {
			return []step.Step{step.OpenPortal()}
		}))
	}

	// Clear ArachnidLair
	actions = append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))

	return
}
