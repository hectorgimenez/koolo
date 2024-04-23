package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Eldritch struct {
	baseRun
}

func (a Eldritch) Name() string {
	return "Eldritch"
}

func (a Eldritch) BuildActions() (actions []action.Action) {
	actions = append(actions,
		a.builder.WayPoint(area.FrigidHighlands), // Moving to starting point (Frigid Highlands)
		a.builder.Buff(),
		a.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			if m, found := d.Monsters.FindOne(npc.MinionExp, data.MonsterTypeSuperUnique); found {
				return m.UnitID, true
			}

			return 0, false
		}, nil),
		a.builder.ItemPickup(false, 35),
	)

	if a.CharacterCfg.Game.Eldritch.KillShenk {
		actions = append(actions,
			a.builder.MoveToCoords(data.Position{X: 3876, Y: 5130}),
			a.char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				if m, found := d.Monsters.FindOne(npc.OverSeer, data.MonsterTypeSuperUnique); found {
					return m.UnitID, true
				}

				return 0, false
			}, nil),
			a.builder.ItemPickup(false, 35),
		)
	}

	return
}
