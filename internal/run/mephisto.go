package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

var mephistoSafePosition = data.Position{
	X: 17568,
	Y: 8069,
}

type Mephisto struct {
	baseRun
}

func (m Mephisto) Name() string {
	return "Mephisto"
}

func (m Mephisto) BuildActions() []action.Action {
	actions := []action.Action{
		m.builder.WayPoint(area.DuranceOfHateLevel2), // Moving to starting point (Durance of Hate Level 2)
		m.builder.MoveToArea(area.DuranceOfHateLevel3),
		m.builder.MoveToCoords(mephistoSafePosition), // Travel to boss position
		m.char.KillMephisto(),                        // Kill Mephisto
	}

	if config.Config.Game.Mephisto.KillCouncilMembers || config.Config.Game.Mephisto.OpenChests {
		actions = append(actions,
			m.builder.ItemPickup(true, 40),
			m.builder.ClearArea(config.Config.Game.Mephisto.OpenChests, func(monsters data.Monsters) []data.Monster {
				councilMembers := make([]data.Monster, 0)
				// Let's skip all the monsters in case we don't want to kill them but open chests
				if !config.Config.Game.Mephisto.KillCouncilMembers {
					return councilMembers
				}

				for _, mo := range monsters {
					if mo.Name == npc.CouncilMember || mo.Name == npc.CouncilMember2 || mo.Name == npc.CouncilMember3 {
						councilMembers = append(councilMembers, mo)
					}
				}

				return councilMembers
			}),
		)
	}

	return actions
}
