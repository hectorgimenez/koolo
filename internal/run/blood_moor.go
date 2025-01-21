package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/config"
)

type BloodMoor struct {
	ctx *context.Status
}

func NewBloodMoor() *BloodMoor {
	return &BloodMoor{
		ctx: context.Get(),
	}
}

func (b BloodMoor) Name() string {
	return string(config.BloodMoorRun)
}

func (b BloodMoor) Run() error {
    // Get the monster filter based on configuration
    monsterFilter := GetMonsterFilter(b)
    
    // Shall we open chests?
    openChests := b.ctx.CharacterCfg.Game.BloodMoor.OpenChests

    // Less issues finding the waypoint than the town's exit
    if err := action.WayPoint(area.ColdPlains); err != nil {
    	return err
    }
	
	// Buff before we start
    action.Buff()
	
	// Moving to the Blood Moor
    if err = action.MoveToArea(area.BloodMoor); err != nil {
    	return err
    }
    
    // If we don't want to just do the Den of Evil, we clear the area
    if !b.ctx.CharacterCfg.Game.BloodMoor.OnlyClearDenOfEvil {
        if err = action.ClearCurrentLevel(openChests, monsterFilter); err != nil {
            return err
        }
    }
    
    // if we don't want to clear the den of evil, the run is over
    if !b.ctx.CharacterCfg.Game.BloodMoor.ClearDenOfEvil {
        return nil
    }
    
    // else, we go to the den of evil and clear it
    if err = action.MoveToArea(area.DenOfEvil); err != nil {
    	return err
    }
    
    // Buff before we start
    action.Buff()
    
    if err = action.ClearCurrentLevel(openChests, monsterFilter); err != nil {
        return err
    }

    return nil
}

func GetMonsterFilter(b BloodMoor) data.MonsterFilter {
    if b.ctx.CharacterCfg.Game.BloodMoor.FocusOnElitePacks {
        return data.MonsterEliteFilter()
    } else {
        return data.MonsterAnyFilter()
    }
}
