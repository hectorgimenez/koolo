package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

type ArachnidLair struct {
	baseRun
}

func (a ArachnidLair) Name() string {
	return string(config.ArachnidLairRun)
}

func (a ArachnidLair) BuildActions() []action.Action {
	actions := []action.Action{
		a.builder.WayPoint(area.SpiderForest), // Moving to starting point (Spider Forest)
		a.builder.MoveToArea(area.SpiderCave), // Travel to ArachnidLair
	}

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	// Clear ArachnidLair
	return append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))
}
