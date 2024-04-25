package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

type AncientTunnels struct {
	baseRun
}

func (a AncientTunnels) Name() string {
	return string(config.AncientTunnelsRun)
}

func (a AncientTunnels) BuildActions() []action.Action {
	openChests := a.CharacterCfg.Game.AncientTunnels.OpenChests
	onlyElites := a.CharacterCfg.Game.AncientTunnels.FocusOnElitePacks
	filter := data.MonsterAnyFilter()

	if onlyElites {
		filter = data.MonsterEliteFilter()
	}

	actions := []action.Action{
		a.builder.WayPoint(area.LostCity),         // Moving to starting point (Lost City)
		a.builder.MoveToArea(area.AncientTunnels), // Travel to ancient tunnels
	}

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	// Clear Ancient Tunnels
	return append(actions, a.builder.ClearArea(openChests, filter))
}
