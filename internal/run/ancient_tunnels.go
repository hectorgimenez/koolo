package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type AncientTunnels struct {
	ctx *context.Status
}

func NewAncientTunnels() *AncientTunnels {
	return &AncientTunnels{
		ctx: context.Get(),
	}
}

func (a AncientTunnels) Name() string {
	return string(config.AncientTunnelsRun)
}

func (a AncientTunnels) Run() error {
	openChests := a.ctx.CharacterCfg.Game.AncientTunnels.OpenChests
	openSuperChests := a.ctx.CharacterCfg.Game.AncientTunnels.OpenSuperChests
	onlyElites := a.ctx.CharacterCfg.Game.AncientTunnels.FocusOnElitePacks
	filter := data.MonsterAnyFilter()

	if onlyElites {
		filter = data.MonsterEliteFilter()
	}

	err := action.WayPoint(area.LostCity) // Moving to starting point (Lost City)
	if err != nil {
		return err
	}

	err = action.MoveToArea(area.AncientTunnels) // Travel to ancient tunnels
	if err != nil {
		return err
	}
	action.OpenTPIfLeader()

	// Clear Ancient Tunnels
	if a.ctx.CharacterCfg.Game.AncientTunnels.OpenSuperChests {
		if err := action.ClearCurrentLevelSuperChest(openSuperChests, filter); err != nil {
			return err
		}
	}

	return action.ClearCurrentLevel(openChests, filter)
}
