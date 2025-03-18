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

	// Use waypoint to crystalinepassage
	err := action.WayPoint(area.CrystallinePassage)
	if err != nil {
		return err
	}
	_ = action.OpenTPIfLeader()
	// Move to ArreatPlateau
	if err = action.MoveToArea(area.ArreatPlateau); err != nil {
		return err
	}
	action.OpenTPIfLeader()
	// Kill Threshsocket
	return t.ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(npc.BloodBringer, data.MonsterTypeSuperUnique); found {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}
