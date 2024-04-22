package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

type StonyTomb struct {
	baseRun
}

func (a StonyTomb) Name() string {
	return "StonyTomb"
}

func (a StonyTomb) BuildActions() (actions []action.Action) {
	openChests := a.CharacterCfg.Game.StonyTomb.OpenChests
	onlyElites := a.CharacterCfg.Game.StonyTomb.FocusOnElitePacks
	filter := data.MonsterAnyFilter()

	if onlyElites {
		filter = data.MonsterEliteFilter()
	}

	actions = append(actions,
		a.builder.WayPoint(area.DryHills),
		a.builder.MoveToArea(area.RockyWaste),
		a.builder.MoveToArea(area.StonyTombLevel1),
	)

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	// Clear both level of Stony Tomb
	actions = append(actions,
		a.builder.ClearArea(openChests, filter),
		a.builder.MoveToArea(area.StonyTombLevel2),
		a.builder.ClearArea(openChests, filter),
	)

	return
}
