package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
)

type ArachnidLair struct {
	baseRun
}

func (a ArachnidLair) Name() string {
	return "ArachnidLair"
}

func (a ArachnidLair) BuildActions() []action.Action {
	actions := []action.Action{
		a.builder.WayPoint(area.SpiderForest), // Moving to starting point (Spider Forest)
		a.builder.MoveToArea(area.SpiderCave), // Travel to ArachnidLair
	}

	if a.CharacterCfg.Companion.Enabled && a.CharacterCfg.Companion.Leader {
		actions = append(actions, action.NewStepChain(func(_ data.Data) []step.Step {
			return []step.Step{step.OpenPortal(a.CharacterCfg.Bindings.TP)}
		}))
	}

	// Clear ArachnidLair
	return append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))
}
