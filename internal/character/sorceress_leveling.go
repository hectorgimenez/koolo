package character

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type SorceressLeveling struct {
	BaseCharacter
}

func (s SorceressLeveling) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s SorceressLeveling) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0

	for {
		id, found := monsterSelector(*s.data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= NovaSorceressMaxAttacksLoop {
			return nil
		}

		monster, found := s.data.Monsters.FindByID(id)
		if !found {
			s.logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		lvl, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
		if s.data.PlayerUnit.MPPercent() < 15 && lvl.Value < 15 {
			s.logger.Debug("Low mana, using primary attack")
			step.PrimaryAttack(id, 1, false, step.Distance(1, 3))
		} else {
			if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Blizzard); found {
				s.logger.Debug("Using Blizzard")
				step.SecondaryAttack(skill.Blizzard, id, 1, step.Distance(25, 30))
			} else if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Meteor); found {
				s.logger.Debug("Using Meteor")
				step.SecondaryAttack(skill.Meteor, id, 1, step.Distance(25, 30))
			} else if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.FireBall); found {
				s.logger.Debug("Using FireBall")
				step.SecondaryAttack(skill.FireBall, id, 4, step.Distance(25, 30))
			} else if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.IceBolt); found {
				s.logger.Debug("Using IceBolt")
				step.SecondaryAttack(skill.IceBolt, id, 4, step.Distance(25, 30))
			} else {
				s.logger.Debug("No secondary skills available, using primary attack")
				step.PrimaryAttack(id, 1, false, step.Distance(1, 3))
			}
		}

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s SorceressLeveling) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s SorceressLeveling) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.FrozenArmor); found {
		skillsList = append(skillsList, skill.FrozenArmor)
	}

	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	s.logger.Info("Buff skills", "skills", skillsList)
	return skillsList
}

func (s SorceressLeveling) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s SorceressLeveling) staticFieldCasts() int {
	casts := 6
	ctx := context.Get()

	switch ctx.CharacterCfg.Game.Difficulty {
	case difficulty.Normal:
		casts = 8
	}
	s.logger.Debug("Static Field casts", "count", casts)
	return casts
}

func (s SorceressLeveling) ShouldResetSkills() bool {
	lvl, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 24 && s.data.PlayerUnit.Skills[skill.FireBall].Level > 1 {
		s.logger.Info("Resetting skills: Level 24+ and FireBall level > 1")
		return true
	}
	return false
}

func (s SorceressLeveling) SkillsToBind() (skill.ID, []skill.ID) {
	level, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
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

	if s.data.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		skillBindings = append(skillBindings, skill.Blizzard)
	} else if s.data.PlayerUnit.Skills[skill.Meteor].Level > 0 {
		skillBindings = append(skillBindings, skill.Meteor)
	} else if s.data.PlayerUnit.Skills[skill.FireBall].Level > 0 {
		skillBindings = append(skillBindings, skill.FireBall)
	} else if s.data.PlayerUnit.Skills[skill.IceBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.IceBolt)
	}

	mainSkill := skill.AttackSkill
	if s.data.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		mainSkill = skill.Blizzard
	} else if s.data.PlayerUnit.Skills[skill.Meteor].Level > 0 {
		mainSkill = skill.Meteor
	}

	s.logger.Info("Skills bound", "mainSkill", mainSkill, "skillBindings", skillBindings)
	return mainSkill, skillBindings
}

func (s SorceressLeveling) StatPoints() map[stat.ID]int {
	lvl, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
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

func (s SorceressLeveling) SkillPoints() []skill.ID {
	lvl, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
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

func (s SorceressLeveling) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAndariel() error {
	m, _ := s.data.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(3, 5))

	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillDuriel() error {
	m, _ := s.data.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := s.pf.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.pf.DistanceFromMe(councilMembers[j].Position)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s SorceressLeveling) KillMephisto() error {
	m, _ := s.data.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillIzual() error {
	m, _ := s.data.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5))
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)

	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.data.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			time.Sleep(200)
			continue
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		_ = step.SecondaryAttack(skill.StaticField, diablo.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

		return s.killMonster(npc.Diablo, data.MonsterTypeNone)
	}
}

func (s SorceressLeveling) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAncients() error {
	for _, m := range s.data.Monsters.Enemies(data.MonsterEliteFilter()) {
		m, _ := s.data.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)

		step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(8, 10))

		step.MoveTo(data.Position{X: 10062, Y: 12639})

		s.killMonster(m.Name, data.MonsterTypeSuperUnique)
	}
	return nil
}

func (s SorceressLeveling) KillBaal() error {
	m, _ := s.data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
	step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 4))
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)

	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
