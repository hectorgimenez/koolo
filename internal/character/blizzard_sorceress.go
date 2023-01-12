package character

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	sorceressMaxAttacksLoop = 10
	sorceressMinDistance    = 25
	sorceressMaxDistance    = 30
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) KillMonsterSequence(
	monsterSelector func(data game.Data) (game.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) *action.DynamicAction {
	completedAttackLoops := 0

	previousUnitID := 0
	return action.BuildDynamic(func(data game.Data) ([]step.Step, bool) {
		id, found := monsterSelector(data)
		if !found {
			return []step.Step{}, false
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(data, id, skipOnImmunities) {
			return []step.Step{}, false
		}

		if len(opts) == 0 {
			opts = append(opts, step.Distance(sorceressMinDistance, sorceressMaxDistance))
		}
		//if useStaticField {
		//	steps = append(steps,
		//		step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, maxDistance)),
		//		step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, id, 5, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, 15)),
		//	)
		//}
		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}, false
		}

		steps := make([]step.Step, 0)
		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player
		for _, m := range data.Monsters.Enemies() {
			if d := pather.DistanceFromMe(data, m.Position); d < 4 {
				s.logger.Debug("Monster detected close to the player, casting Blizzard over it")
				steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, m.UnitID, 1, opts...))
				break
			}
		}

		// In case monster is stuck behind a wall or character is not able to reachh it we will short the distance
		if completedAttackLoops > 5 {
			s.logger.Debug("Looks like monster is not reachable, moving closer")
			opts = append(opts, step.Distance(2, 8))
		}

		steps = append(steps,
			step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, opts...),
			step.PrimaryAttack(id, 4, opts...),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps, true
	})
}

func (s BlizzardSorceress) Buff() action.Action {
	return action.BuildStatic(func(data game.Data) (steps []step.Step) {
		steps = append(steps, s.buffCTA()...)
		steps = append(steps, step.SyncStep(func(data game.Data) error {
			if config.Config.Bindings.Sorceress.FrozenArmor != "" {
				hid.PressKey(config.Config.Bindings.Sorceress.FrozenArmor)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
			}

			return nil
		}))

		return
	})
}

func (s BlizzardSorceress) KillCountess() action.Action {
	return s.killMonsterByName(npc.DarkStalker, game.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillAndariel() action.Action {
	return s.killMonsterByName(npc.Andariel, game.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillSummoner() action.Action {
	return s.killMonsterByName(npc.Summoner, game.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonsterByName(npc.DefiledWarrior, game.MonsterTypeSuperUnique, sorceressMaxDistance, false, skipOnImmunities)
}

func (s BlizzardSorceress) KillMephisto() action.Action {
	return s.killMonsterByName(npc.Mephisto, game.MonsterTypeNone, sorceressMaxDistance, true, nil)
}

func (s BlizzardSorceress) KillNihlathak() action.Action {
	return s.killMonsterByName(npc.Nihlathak, game.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillCouncil() action.Action {
	return s.KillMonsterSequence(func(data game.Data) (game.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []game.Monster
		var coldImmunes []game.Monster
		for _, m := range data.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.ColdImmune) {
					coldImmunes = append(coldImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}
		}

		councilMembers = append(councilMembers, coldImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s BlizzardSorceress) killMonsterByName(id npc.ID, monsterType game.MonsterType, maxDistance int, useStaticField bool, skipOnImmunities []stat.Resist) action.Action {
	return s.KillMonsterSequence(func(data game.Data) (game.UnitID, bool) {
		if m, found := data.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities, step.Distance(sorceressMinDistance, maxDistance))
}
