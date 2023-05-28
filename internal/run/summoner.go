package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
)

type Summoner struct {
	baseRun
}

func (s Summoner) Name() string {
	return "Summoner"
}

func (s Summoner) BuildActions() (actions []action.Action) {
	// Moving to starting point (Arcane Sanctuary)
	actions = append(actions, s.builder.WayPoint(area.ArcaneSanctuary))

	// Buff
	actions = append(actions, s.char.Buff())

	// Travel to boss position
	actions = append(actions, s.builder.MoveTo(func(d data.Data) (data.Position, bool) {
		m, found := d.NPCs.FindOne(npc.Summoner)

		return m.Positions[0], found
	}))

	// Kill Summoner
	actions = append(actions, s.char.KillSummoner())
	return
}
