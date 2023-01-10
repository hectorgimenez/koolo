package run

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"github.com/hectorgimenez/koolo/internal/game/npc"
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

	actions = append(actions, a.char.KillMonsterSequence(func(data game.Data) (game.UnitID, bool) {
		if m, found := data.Monsters.FindOne(npc.MinionExp, game.MonsterTypeSuperUnique); found {
			return m.UnitID, true
		}

		return 0, false
	}, nil))

	return
}
