package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
)

type AncientTunnels struct {
	baseRun
}

func (a AncientTunnels) Name() string {
	return "AncientTunnels"
}

func (a AncientTunnels) BuildActions() []action.Action {
	actions := []action.Action{
		a.builder.WayPoint(area.LostCity),         // Moving to starting point (Lost City)
		a.builder.MoveToArea(area.AncientTunnels), // Travel to ancient tunnels
	}

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	// Clear Ancient Tunnels
	return append(actions, a.builder.ClearArea(true, data.MonsterAnyFilter()))
}
