package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Palace struct {
	ctx *context.Status
}

func NewPalace() *Palace {
	return &Palace{
		ctx: context.Get(),
	}
}

func (a Palace) Name() string {
	return string(config.PalaceRun)
}

func (a Palace) Run() error {
	openChests := a.ctx.CharacterCfg.Game.AncientTunnels.OpenChests
	onlyElites := a.ctx.CharacterCfg.Game.AncientTunnels.FocusOnElitePacks
	filter := data.MonsterAnyFilter()

	if onlyElites {
		filter = data.MonsterEliteFilter()
	}

	err := action.WayPoint(area.PalaceCellarLevel1)
	if err != nil {
		return err
	}

	// Open a TP if we're the leader
	action.OpenTPIfLeader()

	// Buff before we start
	action.Buff()

	err = action.MoveToArea(area.HaremLevel2)
	if err != nil {
		return err
	}

	if err = action.ClearCurrentLevel(openChests, filter); err != nil {
		return err
	}

	action.Buff()

	err = action.MoveToArea(area.PalaceCellarLevel1)
	if err != nil {
		return err
	}

	if err = action.ClearCurrentLevel(openChests, filter); err != nil {
		return err
	}

	action.Buff()

	err = action.MoveToArea(area.PalaceCellarLevel2)
	if err != nil {
		return err
	}

	if err = action.ClearCurrentLevel(openChests, filter); err != nil {
		return err
	}

	action.Buff()

	err = action.MoveToArea(area.PalaceCellarLevel3)
	if err != nil {
		return err
	}

	if err = action.ClearCurrentLevel(openChests, filter); err != nil {
		return err
	}

	// Return to town
	if err = action.ReturnTown(); err != nil {
		return err
	}

	return nil
}
