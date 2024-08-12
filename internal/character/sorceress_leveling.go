package character

import (
	"log/slog"
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

type SorceressLeveling struct {
	BaseCharacter
}

func (s SorceressLeveling) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := d.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s SorceressLeveling) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
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
				s.logger.Debug("Using Blizzard")
				steps = append(steps, step.SecondaryAttack(skill.Blizzard, id, 1, step.Distance(25, 30)))
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.Meteor); found {
				s.logger.Debug("Using Meteor")
				steps = append(steps, step.SecondaryAttack(skill.Meteor, id, 1, step.Distance(25, 30)))
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.FireBall); found {
				s.logger.Debug("Using FireBall")
				steps = append(steps, step.SecondaryAttack(skill.FireBall, id, 4, step.Distance(25, 30)))
			} else if _, found := d.KeyBindings.KeyBindingForSkill(skill.IceBolt); found {
				s.logger.Debug("Using IceBolt")
				steps = append(steps, step.SecondaryAttack(skill.IceBolt, id, 4, step.Distance(25, 30)))
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

func (s SorceressLeveling) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	s.logger.Info("Killing monster", "npc", npc, "type", t)
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s SorceressLeveling) BuffSkills(d game.Data) []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.FrozenArmor); found {
		skillsList = append(skillsList, skill.FrozenArmor)
	}

	if _, found := d.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	s.logger.Info("Buff skills", "skills", skillsList)
	return skillsList
}

func (s SorceressLeveling) PreCTABuffSkills(_ game.Data) []skill.ID {
	return []skill.ID{}
}

func (s SorceressLeveling) staticFieldCasts() int {
	casts := 6
	switch s.container.CharacterCfg.Game.Difficulty {
	case difficulty.Normal:
		casts = 8
	}
	s.logger.Debug("Static Field casts", "count", casts)
	return casts
}

func (s SorceressLeveling) ShouldResetSkills(d game.Data) bool {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 24 && d.PlayerUnit.Skills[skill.FireBall].Level > 1 {
		s.logger.Info("Resetting skills: Level 24+ and FireBall level > 1")
		return true
	}
	return false
}

func (s SorceressLeveling) SkillsToBind(d game.Data) (skill.ID, []skill.ID) {
	level, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	skillBindings := []skill.ID{
		skill.TomeOfTownPortal,
	}

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
	} else if d.PlayerUnit.Skills[skill.Meteor].Level > 0 {
		skillBindings = append(skillBindings, skill.Meteor)
	} else if d.PlayerUnit.Skills[skill.FireBall].Level > 0 {
		skillBindings = append(skillBindings, skill.FireBall)
	} else if d.PlayerUnit.Skills[skill.IceBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.IceBolt)
	}

	mainSkill := skill.AttackSkill
	if d.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		mainSkill = skill.Blizzard
	} else if d.PlayerUnit.Skills[skill.Meteor].Level > 0 {
		mainSkill = skill.Meteor
	}

	s.logger.Info("Skills bound", "mainSkill", mainSkill, "skillBindings", skillBindings)
	return mainSkill, skillBindings
}

func (s SorceressLeveling) StatPoints(d game.Data) map[stat.ID]int {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	statPoints := make(map[stat.ID]int)

	if lvl.Value < 20 {
		statPoints[stat.Vitality] = 9999
	} else {
		statPoints[stat.Energy] = 80
		statPoints[stat.Strength] = 60
		statPoints[stat.Vitality] = 9999
	}

	s.logger.Info("Assigning stat points", "level", lvl.Value, "statPoints", statPoints)
	return statPoints
}

func (s SorceressLeveling) SkillPoints(d game.Data) []skill.ID {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	var skillPoints []skill.ID

	if lvl.Value < 24 {
		skillPoints = []skill.ID{
			skill.FireBolt,
			skill.FireBolt,
			skill.FireBolt,
			skill.FrozenArmor,
			skill.FireBolt,
			skill.StaticField,
			skill.FireBolt,
			skill.Warmth,
			skill.FireBolt,
			skill.Telekinesis,
			skill.FireBolt,
			skill.FireBolt,
			skill.FireBolt,
			skill.FireBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.Teleport,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
		}
	} else {
		skillPoints = []skill.ID{
			skill.FireBolt,
			skill.Warmth,
			skill.Inferno,
			skill.Blaze,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.Meteor,
			skill.FireMastery,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
			skill.Meteor,
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

	s.logger.Info("Assigning skill points", "level", lvl.Value, "skillPoints", skillPoints)
	return skillPoints
}

func (s SorceressLeveling) KillCountess() action.Action {
	s.logger.Info("Starting Countess kill sequence")
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAndariel() action.Action {
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

func (s SorceressLeveling) KillSummoner() action.Action {
	s.logger.Info("Starting Summoner kill sequence")
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillDuriel() action.Action {
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

func (s SorceressLeveling) KillCouncil() action.Action {
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

func (s SorceressLeveling) KillMephisto() action.Action {
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

func (s SorceressLeveling) KillIzual() action.Action {
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

func (s SorceressLeveling) KillDiablo() action.Action {
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

func (s SorceressLeveling) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	s.logger.Info("Starting Pindleskin kill sequence")
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillNihlathak() action.Action {
	s.logger.Info("Starting Nihlathak kill sequence")
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAncients() action.Action {
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

func (s SorceressLeveling) KillBaal() action.Action {
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
