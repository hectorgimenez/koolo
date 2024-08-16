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

type SorceressLevelingLightning struct {
	BaseCharacter
}

func (s SorceressLevelingLightning) CheckKeyBindings(d game.Data) []skill.ID {

	// Not implemented
	return []skill.ID{}
}

func (s SorceressLevelingLightning) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0

	return action.NewStepChain(func(d game.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			s.logger.Debug("No monster found to attack")
			return []step.Step{}
		}
		if previousUnitID != int(id) {
			s.logger.Info("New monster targeted", "id", id)
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(d, id, skipOnImmunities) {
			s.logger.Debug("Pre-battle checks failed")
			return []step.Step{}
		}

		if len(opts) == 0 {
			opts = append(opts, step.Distance(1, 30))
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			s.logger.Info("Max attack loops reached", "loops", sorceressMaxAttacksLoop)
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
		if d.PlayerUnit.MPPercent() < 15 && lvl.Value < 15 {
			s.logger.Debug("Low mana, using primary attack")
			steps = append(steps, step.PrimaryAttack(id, 1, false, step.Distance(1, 3)))
		} else {
			if _, found := d.KeyBindings.KeyBindingForSkill(skill.Blizzard); found {
				if completedAttackLoops%2 == 0 {
					for _, m := range d.Monsters.Enemies() {
						if d := pather.DistanceFromMe(d, m.Position); d < 4 {
							s.logger.Debug("Monster close, casting Blizzard")
							steps = append(steps, step.SecondaryAttack(skill.Blizzard, m.UnitID, 1, step.Distance(25, 30)))
							break
						}
					}
				}

				s.logger.Debug("Using Blizzard")
				steps = append(steps,
					step.SecondaryAttack(skill.Blizzard, id, 1, step.Distance(25, 30)),
					step.PrimaryAttack(id, 3, false, step.Distance(25, 30)),
				)
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.Nova); found {
				s.logger.Debug("Using Nova")
				steps = append(steps, step.SecondaryAttack(skill.Nova, id, 4, step.Distance(1, 5)))
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.ChargedBolt); found {
				s.logger.Debug("Using ChargedBolt")
				steps = append(steps, step.SecondaryAttack(skill.ChargedBolt, id, 4, step.Distance(1, 5)))
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.FireBolt); found {
				s.logger.Debug("Using FireBolt")
				steps = append(steps, step.SecondaryAttack(skill.FireBolt, id, 4, step.Distance(1, 5)))
			} else {
				s.logger.Debug("No secondary skills available, using primary attack")
				steps = append(steps, step.PrimaryAttack(id, 1, false, step.Distance(1, 3)))
			}
		}

		completedAttackLoops++
		previousUnitID = int(id)

		s.logger.Debug("Attack sequence completed", "steps", len(steps), "loops", completedAttackLoops)
		return steps
	}, action.RepeatUntilNoSteps())
}

func (s SorceressLevelingLightning) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	s.logger.Info("Killing monster", "npc", npc, "type", t)
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s SorceressLevelingLightning) BuffSkills(d game.Data) []skill.ID {
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
			break
		}
	}

	return skillsList
}

func (s SorceressLevelingLightning) PreCTABuffSkills(_ game.Data) []skill.ID {
	return []skill.ID{}
}

func (s SorceressLevelingLightning) staticFieldCasts() int {
	casts := 6
	switch s.container.CharacterCfg.Game.Difficulty {
	case difficulty.Normal:
		casts = 8
	}
	s.logger.Debug("Static Field casts", "count", casts)
	return casts
}

func (s SorceressLevelingLightning) ShouldResetSkills(d game.Data) bool {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 25 && d.PlayerUnit.Skills[skill.Nova].Level > 10 {
		s.logger.Info("Resetting skills: Level 25+ and Nova level > 10")
		return true
	}

	return false
}

func (s SorceressLevelingLightning) SkillsToBind(d game.Data) (skill.ID, []skill.ID) {
	level, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	skillBindings := []skill.ID{
		skill.TomeOfTownPortal,
	}

	// Add skills only if the character level is high enough
	if level.Value >= 4 {
		skillBindings = append(skillBindings, skill.FrozenArmor)
	}
	if level.Value >= 6 {
		skillBindings = append(skillBindings, skill.StaticField)
	}
	if level.Value >= 18 {
		skillBindings = append(skillBindings, skill.Teleport)
	}

	if d.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		skillBindings = append(skillBindings, skill.Blizzard)
	} else if d.PlayerUnit.Skills[skill.Nova].Level > 1 {
		skillBindings = append(skillBindings, skill.Nova)
	} else if d.PlayerUnit.Skills[skill.ChargedBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.ChargedBolt)
	} else if d.PlayerUnit.Skills[skill.FireBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.FireBolt)
	}

	mainSkill := skill.AttackSkill
	if d.PlayerUnit.Skills[skill.GlacialSpike].Level > 0 {
		mainSkill = skill.GlacialSpike
	}

	s.logger.Info("Skills bound", "mainSkill", mainSkill, "skillBindings", skillBindings)
	return mainSkill, skillBindings
}

func (s SorceressLevelingLightning) StatPoints(d game.Data) map[stat.ID]int {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	statPoints := make(map[stat.ID]int)

	if lvl.Value < 9 {
		statPoints[stat.Vitality] = 9999
	} else if lvl.Value < 15 {
		statPoints[stat.Energy] = 45
		statPoints[stat.Strength] = 25
		statPoints[stat.Vitality] = 9999
	} else {
		statPoints[stat.Energy] = 60
		statPoints[stat.Strength] = 50
		statPoints[stat.Vitality] = 9999
	}

	s.logger.Info("Assigning stat points", "level", lvl.Value, "statPoints", statPoints)
	return statPoints
}

func (s SorceressLevelingLightning) SkillPoints(d game.Data) []skill.ID {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	var skillPoints []skill.ID

	if lvl.Value < 25 {
		skillPoints = []skill.ID{
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.ChargedBolt,
			skill.FrozenArmor,
			skill.ChargedBolt,
			skill.StaticField,
			skill.StaticField,
			skill.StaticField,
			skill.StaticField,
			skill.Telekinesis,
			skill.Warmth,
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
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
			skill.Nova,
		}
	} else {
		skillPoints = []skill.ID{
			skill.StaticField,
			skill.StaticField,
			skill.StaticField,
			skill.StaticField,
			skill.Telekinesis,
			skill.Teleport,
			skill.FrozenArmor,
			skill.IceBolt,
			skill.IceBlast,
			skill.FrostNova,
			skill.GlacialSpike,
			skill.Blizzard,
			skill.Blizzard,
			skill.Warmth,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.ColdMastery,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.Blizzard,
			skill.ColdMastery,
			skill.ColdMastery,
			skill.ColdMastery,
			skill.ColdMastery,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
		}
	}

	s.logger.Info("Assigning skill points", "level", lvl.Value, "skillPoints", skillPoints)
	return skillPoints
}

func (s SorceressLevelingLightning) KillCountess() action.Action {
	s.logger.Info("Starting Countess kill sequence")
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingLightning) KillAndariel() action.Action {
	s.logger.Info("Starting Andariel kill sequence")
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
				s.logger.Info("Casting Static Field on Andariel")
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(3, 5)),
				}
			}),
			s.killMonster(npc.Andariel, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLevelingLightning) KillSummoner() action.Action {
	s.logger.Info("Starting Summoner kill sequence")
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s SorceressLevelingLightning) KillDuriel() action.Action {
	s.logger.Info("Starting Duriel kill sequence")
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
				s.logger.Info("Casting Static Field on Duriel")
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Duriel, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLevelingLightning) KillCouncil() action.Action {
	s.logger.Info("Starting Council kill sequence")
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := pather.DistanceFromMe(d, councilMembers[i].Position)
			distanceJ := pather.DistanceFromMe(d, councilMembers[j].Position)

			return distanceI < distanceJ
		})

		if len(councilMembers) > 0 {
			s.logger.Debug("Targeting Council member", "id", councilMembers[0].UnitID)
			return councilMembers[0].UnitID, true
		}

		s.logger.Debug("No Council members found")
		return 0, false
	}, nil)
}

func (s SorceressLevelingLightning) KillMephisto() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		s.logger.Info("Starting Mephisto kill sequence")
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				mephisto, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
				s.logger.Info("Casting Static Field on Mephisto")
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, mephisto.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Mephisto, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLevelingLightning) KillIzual() action.Action {
	s.logger.Info("Starting Izual kill sequence")
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				monster, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				s.logger.Info("Casting Static Field on Izual")
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, monster.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
				}
			}),
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

func (s SorceressLevelingLightning) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
			s.logger.Info("Starting Diablo kill sequence")
		}

		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			if diabloFound {
				s.logger.Info("Diablo killed or not found")
				return nil
			}

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

func (s SorceressLevelingLightning) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	s.logger.Info("Starting Pindleskin kill sequence")
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingLightning) KillNihlathak() action.Action {
	s.logger.Info("Starting Nihlathak kill sequence")
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingLightning) KillAncients() action.Action {
	s.logger.Info("Starting Ancients kill sequence")
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
			actions = append(actions,
				action.NewStepChain(func(d game.Data) []step.Step {
					m, _ := d.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)
					s.logger.Info("Targeting Ancient", "name", m.Name)
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

func (s SorceressLevelingLightning) KillBaal() action.Action {
	s.logger.Info("Starting Baal kill sequence")
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				baal, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				s.logger.Info("Casting Static Field on Baal")
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, baal.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
				}
			}),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
		}
	})
}
