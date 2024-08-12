package character

import (
	"sort"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type SorceressLevelingHydraOrb struct {
	BaseCharacter
}

func (s SorceressLevelingHydraOrb) CheckKeyBindings(d game.Data) []skill.ID {

	// Not implemented
	return []skill.ID{}
}

func (s SorceressLevelingHydraOrb) ShouldResetSkills(d game.Data) bool {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 25 && d.PlayerUnit.Skills[skill.Nova].Level > 10 {
		return true
	}

	return false
}

func (s SorceressLevelingHydraOrb) SkillsToBind(d game.Data) (skill.ID, []skill.ID) {
	skillBindings := []skill.ID{
		skill.FrozenArmor,
		skill.StaticField,
		skill.Teleport,
		skill.TomeOfTownPortal,
	}

	if d.PlayerUnit.Skills[skill.ChargedBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.ChargedBolt)
	} else if d.PlayerUnit.Skills[skill.FireBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.FireBolt)
	}

	mainSkill := skill.AttackSkill
//	if d.PlayerUnit.Skills[skill.GlacialSpike].Level > 0 {
//		mainSkill = skill.GlacialSpike
//	}

	return mainSkill, skillBindings
}

func (s SorceressLevelingHydraOrb) StatPoints(d game.Data) map[stat.ID]int {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value < 5 {
		return map[stat.ID]int{
			stat.Vitality: 9999,
		}
	}

	if lvl.Value < 15 {
		return map[stat.ID]int{
			stat.Energy:   45,
			stat.Strength: 25,
			stat.Vitality: 9999,
		}
	}

	return map[stat.ID]int{
		stat.Energy:   60,
		stat.Strength: 50,
		stat.Vitality: 9999,
	}
}

func (s SorceressLevelingHydraOrb) SkillPoints(d game.Data) []skill.ID {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value < 30 {
		return []skill.ID{
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.StaticField,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,	// one extra from quest
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,			
			skill.Telekinesis,
			skill.Teleport,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
		}
	}

	return []skill.ID{
		skill.IceBolt,
		skill.IceBolt,
		skill.FrozenArmor,
		skill.StaticField,
		skill.FrostNova,
		skill.Warmth,
		skill.Telekinesis,
		skill.Teleport,	
		skill.GlacialSpike,
		skill.Blizzard,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.FrozenOrb,
		skill.IceBolt,
		skill.IceBolt,
		skill.IceBolt,
		skill.IceBolt,
		skill.IceBolt,
		skill.IceBolt,
		skill.IceBolt,
		skill.IceBolt,	
		skill.ColdMastery,
		skill.FireBolt,
		skill.FireBall,
		skill.Enchant,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.FireMastery,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.Hydra,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,
		skill.FireMastery,

	}
}

func (s SorceressLevelingHydraOrb) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingHydraOrb) KillAndariel() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(3, 5)),
				}
			}),
			s.killMonster(npc.Andariel, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLevelingHydraOrb) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s SorceressLevelingHydraOrb) KillDuriel() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Duriel, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLevelingHydraOrb) KillMephisto() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		// Let's try to moat trick if Teleport is available
		//if step.CanTeleport(d) {
		//	moatTrickPosition := data.Position{X: 17611, Y: 8093}
		//	return []action.Action{
		//		action.NewStepChain(func(d game.Data) []step.Step {
		//			mephisto, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
		//			return []step.Step{
		//				step.Wait(time.Second * 2),
		//				step.MoveTo(data.Position{X: 17580, Y: 8085}),
		//				step.Wait(time.Second * 3),
		//				step.MoveTo(moatTrickPosition),
		//				step.Wait(time.Second * 3),
		//				step.SecondaryAttack(s.container.CharacterCfg.Bindings.Sorceress.Blizzard, mephisto.UnitID, 3),
		//			}
		//		}),
		//	}
		//}

		// If teleport is not available, just try to kill him with Static Field and Fire Ball
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				mephisto, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, mephisto.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Mephisto, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLevelingHydraOrb) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingHydraOrb) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingHydraOrb) KillCouncil() action.Action {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := pather.DistanceFromMe(d, councilMembers[i].Position)
			distanceJ := pather.DistanceFromMe(d, councilMembers[j].Position)

			return distanceI < distanceJ
		})

		if len(councilMembers) > 0 {
			return councilMembers[0].UnitID, true
		}

		return 0, false
	}, nil)
}

func (s SorceressLevelingHydraOrb) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
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
			return []action.Action{action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			})}
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, diablo.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s SorceressLevelingHydraOrb) KillIzual() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				monster, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, monster.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
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

func (s SorceressLevelingHydraOrb) KillBaal() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				baal, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, baal.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
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

func (s SorceressLevelingHydraOrb) KillAncients() action.Action {
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
			actions = append(actions,
				action.NewStepChain(func(d game.Data) []step.Step {
					m, _ := d.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)
					return []step.Step{
						step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(8, 10)),
						step.MoveTo(data.Position{
							X: 10062,
							Y: 12639,
						}),
					}
				}),
				s.killMonster(m.Name, data.MonsterTypeSuperUnique),
			)
		}
		return actions
	})
}

func (s SorceressLevelingHydraOrb) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0

	return action.NewStepChain(func(d game.Data) []step.Step {
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
			opts = append(opts, step.Distance(1, 30))
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		// During early game stages amount of mana is ridiculous...
		lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
		if d.PlayerUnit.MPPercent() < 15 && lvl.Value < 15 {
			steps = append(steps, step.PrimaryAttack(id, 1, false, step.Distance(1, 3)))
		} else {
			if _, found := d.KeyBindings.KeyBindingForSkill(skill.Hydra); found {
				if completedAttackLoops%2 == 0 {
					for _, m := range d.Monsters.Enemies() {
						if d := pather.DistanceFromMe(d, m.Position); d > 5 {
							s.logger.Debug("Monster detected at range of the player, casting Hydras")
							steps = append(steps, step.SecondaryAttack(skill.Hydra, m.UnitID, 2, step.Distance(25, 30)))
							break
						}
					}
				}

				steps = append(steps,
					step.SecondaryAttack(skill.Hydra, id, 1, step.Distance(25, 30)),
					step.PrimaryAttack(id, 3, false, step.Distance(25, 30)),
				)
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.Nova); found {
				steps = append(steps, step.SecondaryAttack(skill.Nova, id, 4, step.Distance(1, 5)))
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.ChargedBolt); found {
				steps = append(steps, step.SecondaryAttack(skill.ChargedBolt, id, 4, step.Distance(1, 5)))
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.FireBolt); found {
				steps = append(steps, step.SecondaryAttack(skill.FireBolt, id, 4, step.Distance(1, 5)))
			}
		}

		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s SorceressLevelingHydraOrb) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s SorceressLevelingHydraOrb) BuffSkills(d game.Data) []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	if _, found := d.KeyBindings.KeyBindingForSkill(skill.ThunderStorm); found {
		skillsList = append(skillsList, skill.ThunderStorm)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := d.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			return skillsList
		}
	}

	return skillsList
}

func (s SorceressLevelingHydraOrb) PreCTABuffSkills(_ game.Data) []skill.ID {
	return []skill.ID{}
}

func (s SorceressLevelingHydraOrb) staticFieldCasts() int {
	switch s.container.CharacterCfg.Game.Difficulty {
	case difficulty.Normal:
		return 8
	}

	return 6
}
