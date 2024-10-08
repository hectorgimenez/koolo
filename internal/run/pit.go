package run

import (
	"fmt"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
)

var _ AreaAwareRun = (*Pit)(nil)

type Pit struct {
	ctx *context.Status
}

func NewPit() *Pit {
	return &Pit{
		ctx: context.Get(),
	}
}

func (p *Pit) Name() string {
	return string(config.PitRun)
}

func (p *Pit) Run() error {
	// Use waypoint to Black Marsh
	err := action.WayPoint(area.BlackMarsh)
	if err != nil {
		return fmt.Errorf("error using waypoint to Black Marsh: %w", err)
	}

	// Move through areas
	if p.ctx.CharacterCfg.Game.Pit.MoveThroughBlackMarsh {
		if err := action.MoveToArea(area.TamoeHighland); err != nil {
			return fmt.Errorf("error moving through Black Marsh: %w", err)
		}
	}

	if err := action.MoveToArea(area.PitLevel1); err != nil {
		return fmt.Errorf("error moving to Pit: %w", err)
	}

	// Clear Pit Level 1 if not only clearing Level 2
	if !p.ctx.CharacterCfg.Game.Pit.OnlyClearLevel2 {
		if err := p.clearLevel(area.PitLevel1); err != nil {
			return err
		}
	}

	// Move to and clear Pit Level 2
	if err := action.MoveToArea(area.PitLevel2); err != nil {
		return fmt.Errorf("error moving to Pit Level 2: %w", err)
	}

	return p.clearLevel(area.PitLevel2)
}

func (p *Pit) clearLevel(a area.ID) error {
	monsterFilter := data.MonsterAnyFilter()
	if p.ctx.CharacterCfg.Game.Pit.FocusOnElitePacks {
		monsterFilter = data.MonsterEliteFilter()
	}
	return action.ClearCurrentLevel(p.ctx.CharacterCfg.Game.Pit.OpenChests, monsterFilter)
}

func (p *Pit) ExpectedAreas() []area.ID {
	return []area.ID{
		area.BlackMarsh,
		area.TamoeHighland,
		area.PitLevel1,
		area.PitLevel2,
	}
}

func (p *Pit) IsAreaPartOfRun(a area.ID) bool {
	for _, expectedArea := range p.ExpectedAreas() {
		if a == expectedArea {
			return true
		}
	}
	return false
}
