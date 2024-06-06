package run

import (
	"time"

	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/object"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/pather"
)

var baalThronePosition = data.Position{
	X: 15095,
	Y: 5042,
}

type Baal struct {
	baseRun
}

func (s Baal) Name() string {
	return string(config.BaalRun)
}

func (s Baal) BuildActions() (actions []action.Action) {
	actions = append(actions,
		// Moving to starting point (The World StoneKeep Level 2)
		s.builder.WayPoint(area.TheWorldStoneKeepLevel2),
		// Travel to boss position
		s.builder.MoveToArea(area.TheWorldStoneKeepLevel3),
		s.builder.MoveToArea(area.ThroneOfDestruction),
		s.builder.MoveToCoords(baalThronePosition),
		// Kill monsters inside Baal throne
		s.checkForSoulsOrDolls(),
		s.builder.ClearAreaAroundPlayer(50, data.MonsterAnyFilter()),
	)

	// Let's move to a safe area and open the portal in companion mode
	if s.CharacterCfg.Companion.Enabled && s.CharacterCfg.Companion.Leader {
		actions = append(actions,
			s.builder.MoveToCoords(data.Position{
				X: 15116,
				Y: 5071,
			}),
			s.builder.OpenTPIfLeader(),
		)
	}

	// Come back to previous position
	actions = append(actions, s.builder.Buff(), s.builder.MoveToCoords(baalThronePosition))

	lastWave := false
	actions = append(actions, action.NewChain(func(d game.Data) []action.Action {
		if !lastWave {
			if _, found := d.Monsters.FindOne(npc.BaalsMinion, data.MonsterTypeMinion); found {
				lastWave = true
			}

			enemies := false
			for _, e := range d.Monsters.Enemies() {
				dist := pather.DistanceFromPoint(baalThronePosition, e.Position)
				if dist < 50 {
					enemies = true
				}
			}
			if !enemies {
				return []action.Action{
					s.builder.ItemPickup(false, 50),
					s.builder.MoveToCoords(baalThronePosition),
				}
			}

			return []action.Action{s.builder.ClearAreaAroundPlayer(50, data.MonsterAnyFilter())}
		}

		return nil
	}, action.RepeatUntilNoSteps()))

	actions = append(actions, s.builder.ItemPickup(false, 30))

	_, isLevelingChar := s.char.(action.LevelingCharacter)
	if s.CharacterCfg.Game.Baal.KillBaal || isLevelingChar {
		actions = append(actions,
			s.builder.Wait(time.Second*10),
			s.builder.Buff(),
			s.builder.InteractObject(object.BaalsPortal, func(d game.Data) bool {
				return d.PlayerUnit.Area == area.TheWorldstoneChamber
			}),
			s.char.KillBaal(),
			s.builder.ItemPickup(true, 50),
		)
	}

	return
}

func (s Baal) checkForSoulsOrDolls() *action.Chain {
	return action.NewChain(func(d game.Data) []action.Action {
		var npcIds []npc.ID

		if s.CharacterCfg.Game.Baal.DollQuit {
			npcIds = append(npcIds, npc.UndeadStygianDoll2, npc.UndeadSoulKiller2)
		}
		if s.CharacterCfg.Game.Baal.SoulQuit {
			npcIds = append(npcIds, npc.BlackSoul2, npc.BurningSoul2)
		}

		if len(npcIds) != 0 {
			return []action.Action{
				s.builder.ChickenOnMonsters(60, npcIds),
			}
		}

		return []action.Action{}
	})
}
