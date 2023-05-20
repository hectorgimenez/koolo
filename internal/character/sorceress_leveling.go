package character

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type SorceressLeveling struct {
	BaseCharacter
}

func (s SorceressLeveling) StatPoints() map[stat.ID]int {
	return map[stat.ID]int{
		stat.Dexterity: 0,
		stat.Energy:    0,
		stat.Strength:  15,
		stat.Vitality:  9999,
	}
}

func (s SorceressLeveling) SkillPoints() []skill.Skill {
	return []skill.Skill{
		skill.ChargedBolt,
		skill.ChargedBolt,
		skill.ChargedBolt,
		skill.ChargedBolt,
		skill.FrozenArmor,
		skill.StaticField,
		skill.ChargedBolt,
		skill.StaticField,
		skill.StaticField,
		skill.StaticField,
		skill.Telekinesis,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Teleport,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
	}
}

func (s SorceressLeveling) EnsureStatPoints() action.Action {
	return action.BuildStatic(func(d data.Data) []step.Step {
		_, found := d.PlayerUnit.Stats[stat.StatPoints]
		if !found {
			return []step.Step{}
		}

		return nil
	})
}

func (s SorceressLeveling) EnsureSkillPoints() action.Action {
	return action.BuildStatic(func(d data.Data) []step.Step {
		_, found := d.PlayerUnit.Stats[stat.SkillPoints]
		if !found {
			return []step.Step{}
		}

		return nil
	})
}

func (s SorceressLeveling) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillCouncil() action.Action {
	panic("implement me")
}

func (s SorceressLeveling) KillMonsterSequence(monsterSelector func(d data.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) *action.DynamicAction {
	completedAttackLoops := 0
	previousUnitID := 0
	staticFieldUsed := false

	return action.BuildDynamic(func(d data.Data) ([]step.Step, bool) {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}, false
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}, false
		}

		if len(opts) == 0 {
			opts = append(opts, step.Distance(1, 25))
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}, false
		}

		steps := make([]step.Step, 0)

		// Try to static field when monsters around character are > 5 or elite enemy is close enough
		numberOfCloseMonsters := 0
		eliteFound := false
		for _, m := range d.Monsters.Enemies() {
			if pather.DistanceFromMe(d, m.Position) < 5 {
				numberOfCloseMonsters++
				if m.IsElite() {
					eliteFound = true
				}
			}
		}
		if !staticFieldUsed && (numberOfCloseMonsters > 5 || eliteFound) {
			staticFieldUsed = true
			if d.PlayerUnit.Skills[skill.StaticField] > 0 {
				steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, id, 4, step.Distance(1, 5)))
			}
		}

		// During early game stages amount of mana is ridiculous...
		if d.PlayerUnit.MPPercent() < 15 && d.PlayerUnit.Stats[stat.Level] < 15 {
			steps = append(steps, step.PrimaryAttack(id, 1, step.Distance(1, 3)))
		} else {
			steps = append(steps,
				step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, step.Distance(1, 7)),
			)
		}

		completedAttackLoops++
		previousUnitID = int(id)

		return steps, true
	})
}

func (s SorceressLeveling) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}
		helper.Sleep(100)
		for i := 0; i < 10; i++ {
			steps = append(steps,
				step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, m.UnitID, 1, step.Distance(1, 7)),
			)
		}

		return
	}, action.CanBeSkipped())
}

func (s SorceressLeveling) Buff() action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
		return []step.Step{
			step.SyncStep(func(d data.Data) error {
				if _, found := d.PlayerUnit.Skills[skill.FrozenArmor]; !found {
					return nil
				}

				if config.Config.Bindings.Sorceress.FrozenArmor != "" {
					hid.PressKey(config.Config.Bindings.Sorceress.FrozenArmor)
					helper.Sleep(100)
					hid.Click(hid.RightButton)
				}

				return nil
			}),
		}
	})
}
