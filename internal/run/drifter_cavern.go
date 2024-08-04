package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

type DrifterCavern struct {
	baseRun
}

func (a DrifterCavern) Name() string {
	return string(config.DrifterCavernRun)
}

func (a DrifterCavern) BuildActions() (actions []action.Action) {
	openChests := a.CharacterCfg.Game.DrifterCavern.OpenChests
	onlyElites := a.CharacterCfg.Game.DrifterCavern.FocusOnElitePacks
	filter := data.MonsterAnyFilter()

	if onlyElites {
		filter = data.MonsterEliteFilter()
	}

	actions = []action.Action{
		a.builder.WayPoint(area.GlacialTrail),
		a.builder.MoveToArea(area.DrifterCavern),
	}

	/*actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)*/

	return append(actions, a.builder.ClearArea(openChests, filter))
}
