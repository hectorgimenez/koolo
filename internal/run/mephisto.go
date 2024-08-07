package run

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
)

var mephistoSafePosition = data.Position{
	X: 17568,
	Y: 8069,
}

var mephistoAttackPosition = data.Position{
	X: 17572,
	Y: 8070,
}

var mephistoSaferAttackPosition = data.Position{
	X: 17564,
	Y: 8084,
}

type Mephisto struct {
	baseRun
}

func (m Mephisto) Name() string {
	return string(config.MephistoRun)
}

func (m Mephisto) BuildActions() []action.Action {
	actions := []action.Action{
		m.builder.WayPoint(area.DuranceOfHateLevel2),
		m.builder.MoveToArea(area.DuranceOfHateLevel3),
		m.builder.MoveToCoords(mephistoSafePosition),        // move to starting position
		m.builder.MoveToCoords(mephistoAttackPosition),      // move to attack position
		m.builder.Wait(time.Second * 1),                     // wait 1 second for Mephisto to move
		m.builder.MoveToCoords(mephistoSaferAttackPosition), // move to safer attack position
		m.char.KillMephisto(),                               // attack mephisto
		m.builder.ItemPickup(true, 40),                     // making sure we pick up stuff before moving to A4 through red portal for faster next game town movement
		m.builder.InteractObject(object.HellGate, func(d game.Data) bool {
			return d.PlayerUnit.Area == area.ThePandemoniumFortress
		}),
	}

	if m.CharacterCfg.Game.Mephisto.KillCouncilMembers || m.CharacterCfg.Game.Mephisto.OpenChests {
		actions = append(actions,
			m.builder.ItemPickup(true, 40),
			m.builder.ClearArea(m.CharacterCfg.Game.Mephisto.OpenChests, func(monsters data.Monsters) []data.Monster {
				councilMembers := make([]data.Monster, 0)
				// Let's skip all the monsters in case we don't want to kill them but open chests
				if !m.CharacterCfg.Game.Mephisto.KillCouncilMembers {
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
