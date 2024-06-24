package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

type Mausoleum struct {
	baseRun
}

func (a Mausoleum) Name() string {
	return string(config.MausoleumRun)
}

func (a Mausoleum) BuildActions() []action.Action {
	openChests := a.CharacterCfg.Game.Mausoleum.OpenChests
	onlyElites := a.CharacterCfg.Game.Mausoleum.FocusOnElitePacks
	filter := data.MonsterAnyFilter()

	if onlyElites {
		filter = data.MonsterEliteFilter()
	}

	actions := []action.Action{
		a.builder.WayPoint(area.ColdPlains),
		a.builder.MoveToArea(area.BurialGrounds),
		a.builder.MoveToArea(area.Mausoleum),
	}

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	// Clear Mausoleum
	return append(actions, a.builder.ClearArea(openChests, filter))
}
