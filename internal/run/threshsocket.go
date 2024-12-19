package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Threshsocket struct {
	ctx *context.Status
}

func NewThreshsocket() *Threshsocket {
	return &Threshsocket{
		ctx: context.Get(),
	}
}

func (t Threshsocket) Name() string {
	return string(config.ThreshsocketRun)
}

func (t Threshsocket) Run() error {
	ctx := context.Get()

	// Use waypoint to crystalinepassage
	err := action.WayPoint(area.CrystallinePassage)
	if err != nil {
		return err
	}

	// Move to ArreatPlateau
	if err = action.MoveToArea(area.ArreatPlateau); err != nil {
		return err
	}

	// Kill Threshsocket
	return t.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(npc.BloodBringer, data.MonsterTypeSuperUnique); found {
			monsterIsImmune := false
			for _, resist := range ctx.Data.CharacterCfg.Runtime.ImmunityFilter {
				if m.IsImmune(resist) {
					monsterIsImmune = true
					break
				}
			}

			if monsterIsImmune {
				return 0, false
			}

			return m.UnitID, true
		}

		return 0, false
	})
}
