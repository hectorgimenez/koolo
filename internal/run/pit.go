package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Pit struct {
	ctx *context.Status
}

func NewPit() *Pit {
	return &Pit{
		ctx: context.Get(),
	}
}

func (p Pit) Name() string {
	return string(config.PitRun)
}

func (p Pit) Run() error {

	// Define a default filter
	monsterFilter := data.MonsterAnyFilter()

	// Update filter if we selected to clear only elites
	if p.ctx.CharacterCfg.Game.Pit.FocusOnElitePacks {
		monsterFilter = data.MonsterEliteFilter()
	}

	if !p.ctx.CharacterCfg.Game.Pit.MoveThroughBlackMarsh {
		// Use the waypoint to travel to the OuterCloister
		err := action.WayPoint(area.OuterCloister)
		if err != nil {
			return err
		}

		// Move to the Monastery Gate
		if err = action.MoveToArea(area.MonasteryGate); err != nil {
			return err
		}

		// Move to the Tamoe Highland
		if err = action.MoveToArea(area.TamoeHighland); err != nil {
			return err
		}
	} else {
		// Use the waypoint to travel to the BlackMarsh
		err := action.WayPoint(area.BlackMarsh)
		if err != nil {
			return err
		}

		// Move to the TamoeHighland
		if err = action.MoveToArea(area.TamoeHighland); err != nil {
			return err
		}
	}

	// Travel to pit level 1
	if err := action.MoveToArea(area.PitLevel1); err != nil {
		return err
	}

	// Open a TP If we're the leader
	action.OpenTPIfLeader()

	// Clear the area if we don't have only clear lvl2 selected
	if !p.ctx.CharacterCfg.Game.Pit.OnlyClearLevel2 {
		if err := action.ClearCurrentLevel(p.ctx.CharacterCfg.Game.Pit.OpenChests, monsterFilter); err != nil {
			return err
		}
	}

	// Move to PitLvl2
	if err := action.MoveToArea(area.PitLevel2); err != nil {
		return err
	}

	// Clear it
	return action.ClearCurrentLevel(p.ctx.CharacterCfg.Game.Pit.OpenChests, monsterFilter)
}
