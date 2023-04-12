package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
)

type Eldritch struct {
	baseRun
}

func (a Eldritch) Name() string {
	return "Eldritch"
}

func (a Eldritch) BuildActions() (actions []action.Action) {
	// Moving to starting point (Frigid Highlands)
	actions = append(actions, a.builder.WayPoint(area.FrigidHighlands))

	// Buff
	actions = append(actions, a.char.Buff())

	actions = append(actions, a.char.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(npc.MinionExp, data.MonsterTypeSuperUnique); found {
			return m.UnitID, true
		}

		return 0, false
	}, nil))

	return
}
