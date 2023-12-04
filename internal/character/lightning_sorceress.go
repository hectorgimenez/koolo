package character

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	lightningSorceressMaxAttacksLoop = 10
	lightningSorceressMinDistance    = 2
	lightningSorceressMaxDistance    = 8
)

type LightningSorceress struct {
	BaseCharacter
}

func (s LightningSorceress) KillMonsterSequence(
	monsterSelector func(d data.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0

	return action.NewStepChain(func(d data.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		if len(opts) == 0 {
			opts = append(opts, step.Distance(lightningSorceressMinDistance, lightningSorceressMaxDistance))
		}

		if completedAttackLoops >= lightningSorceressMaxAttacksLoop {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)
		for _, m := range d.Monsters.Enemies() {
			if d := pather.DistanceFromMe(d, m.Position); d < 5 {
				s.logger.Debug("Monster detected close to the player, casting Nova over it")
				steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.Nova, 0, 3, opts...))
				break
			}
		}

		// In case monster is stuck behind a wall or character is not able to reach it we will short the distance
		if completedAttackLoops > 5 {
			if completedAttackLoops == 6 {
				s.logger.Debug("Looks like monster is not reachable, reducing max attack distance.")
			}
			opts = []step.AttackOption{step.Distance(0, 1)}
		}

		steps = append(steps,
			step.SecondaryAttack(config.Config.Bindings.Sorceress.Nova, id, 5, opts...),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s LightningSorceress) BuffSkills() map[skill.Skill]string {
	return map[skill.Skill]string{
		skill.FrozenArmor:  config.Config.Bindings.Sorceress.FrozenArmor,
		skill.EnergyShield: config.Config.Bindings.Sorceress.EnergyShield,
	}
}

func (s LightningSorceress) KillCountess() action.Action {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, lightningSorceressMaxDistance, false, nil)
}

func (s LightningSorceress) KillAndariel() action.Action {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeNone, lightningSorceressMaxDistance, false, nil)
}

func (s LightningSorceress) KillSummoner() action.Action {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeNone, lightningSorceressMaxDistance, false, nil)
}

func (s LightningSorceress) KillDuriel() action.Action {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeNone, lightningSorceressMaxDistance, true, nil)
}

func (s LightningSorceress) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, lightningSorceressMaxDistance, false, skipOnImmunities)
}

func (s LightningSorceress) KillMephisto() action.Action {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeNone, lightningSorceressMaxDistance, true, nil)
}

func (s LightningSorceress) KillNihlathak() action.Action {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, lightningSorceressMaxDistance, false, nil)
}

func (s LightningSorceress) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d data.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			return []action.Action{action.NewStepChain(func(d data.Data) []step.Step {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			})}
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, diablo.UnitID, 5, step.Distance(3, 8)),
				}
			}),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s LightningSorceress) KillIzual() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, m.UnitID, 7, step.Distance(5, 8)),
				}
			}),
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
		}
	})
}

func (s LightningSorceress) KillBaal() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, m.UnitID, 5, step.Distance(5, 8)),
				}
			}),
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
		}
	})
}

func (s LightningSorceress) KillCouncil() action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var coldImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
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
	}, nil, step.Distance(8, lightningSorceressMaxDistance))
}

func (s LightningSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, maxDistance int, useStaticField bool, skipOnImmunities []stat.Resist) action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities, step.Distance(lightningSorceressMinDistance, maxDistance))
}

func (s LightningSorceress) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}
