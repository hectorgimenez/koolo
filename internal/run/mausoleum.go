package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	action2 "github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

type Mausoleum struct {
	ctx *context.Status
}

func NewMausoleum() *Mausoleum {
	return &Mausoleum{
		ctx: context.Get(),
	}
}

func (a Mausoleum) Name() string {
	return string(config.MausoleumRun)
}

func (a Mausoleum) Run() error {

	// Define a defaut filter
	monsterFilter := data.MonsterAnyFilter()

	// Update filter if we selected to clear only elites
	if a.ctx.CharacterCfg.Game.Mausoleum.FocusOnElitePacks {
		monsterFilter = data.MonsterEliteFilter()
	}

	// Use the waypoint
	err := action2.WayPoint(area.ColdPlains)
	if err != nil {
		return err
	}

	// Move to the BurialGrounds
	if err = action2.MoveToArea(area.BurialGrounds); err != nil {
		return err
	}

	// Move to the Mausoleum
	if err = action2.MoveToArea(area.Mausoleum); err != nil {
		return err
	}

	// Open a TP If we're the leader
	action2.OpenTPIfLeader()

	// Clear the area
	return action2.ClearCurrentLevel(a.ctx.CharacterCfg.Game.Mausoleum.OpenChests, monsterFilter)
}
