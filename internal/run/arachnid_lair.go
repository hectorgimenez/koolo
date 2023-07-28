package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
)

type ArachnidLair struct {
	baseRun
}

func (a ArachnidLair) Name() string {
	return "ArachnidLair"
}

func (a ArachnidLair) BuildActions() (actions []action.Action) {
	// Moving to starting point (Lost City)
	actions = append(actions, a.builder.WayPoint(area.SpiderForest))

	// Buff
	actions = append(actions, a.char.Buff())

	// Travel to ArachnidLair
	actions = append(actions, action.BuildStatic(func(d data.Data) []step.Step {
		return []step.Step{
			step.MoveToLevel(area.SpiderCave),
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

	// Clear ArachnidLair
	actions = append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))

	return
}
