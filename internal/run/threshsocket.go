package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Threshsocket struct {
	baseRun
}

func (a Threshsocket) Name() string {
	return string(config.ThreshsocketRun)
}

func (a Threshsocket) BuildActions() (actions []action.Action) {
	return []action.Action{
		a.builder.WayPoint(area.CrystallinePassage), // Moving to starting point (Crystalline Passage)
		a.builder.MoveToArea(area.ArreatPlateau),
		a.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			if m, found := d.Monsters.FindOne(npc.BloodBringer, data.MonsterTypeSuperUnique); found {
				return m.UnitID, true
			}
      

			return 0, false
		}, nil),
		a.builder.ItemPickup(false, 35),
		
	}


}
