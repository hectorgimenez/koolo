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
	ctx            *context.Status
	visitedAreas   map[area.ID]bool
	lastActionArea area.ID
}

func NewPit() *Pit {
	return &Pit{
		ctx:          context.Get(),
		visitedAreas: make(map[area.ID]bool),
	}
}

func (p *Pit) Name() string {
	return string(config.PitRun)
}

func (p *Pit) Run() error {
	expectedAreas := p.ExpectedAreas()

	for _, currentArea := range expectedAreas {
		if !p.visitedAreas[currentArea] {
			if err := p.moveToArea(currentArea); err != nil {
				return err
			}
		}

		p.visitedAreas[currentArea] = true
		p.lastActionArea = currentArea

		if currentArea == area.PitLevel1 && !p.ctx.CharacterCfg.Game.Pit.OnlyClearLevel2 {
			if err := p.clearLevel(area.PitLevel1); err != nil {
				return err
			}
		}

		if currentArea == area.PitLevel2 {
			return p.clearLevel(area.PitLevel2)
		}
	}

	return nil
}

func (p *Pit) moveToArea(targetArea area.ID) error {

	var err error
	if targetArea == p.ExpectedAreas()[0] {
		// Use waypoint for the first area
		err = action.WayPoint(targetArea)
	} else {
		err = action.MoveToArea(targetArea)
	}

	if err != nil {
		return fmt.Errorf("error moving to %s: %w", area.Areas[targetArea].Name, err)
	}
	
	return nil
}

func (p *Pit) clearLevel(a area.ID) error {
	monsterFilter := data.MonsterAnyFilter()
	if p.ctx.CharacterCfg.Game.Pit.FocusOnElitePacks {
		monsterFilter = data.MonsterEliteFilter()
	}
	return action.ClearCurrentLevel(p.ctx.CharacterCfg.Game.Pit.OpenChests, monsterFilter)
}

func (p *Pit) ExpectedAreas() []area.ID {
	if p.ctx.CharacterCfg.Game.Pit.MoveThroughBlackMarsh {
		return []area.ID{
			area.BlackMarsh,
			area.TamoeHighland,
			area.PitLevel1,
			area.PitLevel2,
		}
	}
	return []area.ID{
		area.OuterCloister,
		area.MonasteryGate,
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

func (p *Pit) GetVisitedAreas() map[area.ID]bool {
	return p.visitedAreas
}

func (p *Pit) GetLastActionArea() area.ID {
	return p.lastActionArea
}

func (p *Pit) SetVisitedAreas(areas map[area.ID]bool) {
	p.visitedAreas = areas
}

func (p *Pit) SetLastActionArea(area area.ID) {
	p.lastActionArea = area
}
