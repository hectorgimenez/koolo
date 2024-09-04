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

const (
	SorceressLevelingLightningMaxAttacksLoop = 10
)

type SorceressLevelingLightning struct {
	BaseCharacter
}

func (s SorceressLevelingLightning) CheckKeyBindings() []skill.ID {
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

func (s SorceressLevelingLightning) KillMonsterSequence(
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

		if completedAttackLoops >= SorceressLevelingLightningMaxAttacksLoop {
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
				if completedAttackLoops%2 == 0 {
					for _, m := range s.data.Monsters.Enemies() {
						if d := s.pf.DistanceFromMe(m.Position); d < 4 {
							s.logger.Debug("Monster close, casting Blizzard")
							step.SecondaryAttack(skill.Blizzard, m.UnitID, 1, step.Distance(25, 30))
							break
						}
					}
				}

				s.logger.Debug("Using Blizzard")

				step.SecondaryAttack(skill.Blizzard, id, 1, step.Distance(25, 30))
				step.PrimaryAttack(id, 3, false, step.Distance(25, 30))

			} else if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Nova); found {
				s.logger.Debug("Using Nova")
				step.SecondaryAttack(skill.Nova, id, 4, step.Distance(1, 5))
			} else if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.ChargedBolt); found {
				s.logger.Debug("Using ChargedBolt")
				step.SecondaryAttack(skill.ChargedBolt, id, 4, step.Distance(1, 5))
			} else if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.FireBolt); found {
				s.logger.Debug("Using FireBolt")
				step.SecondaryAttack(skill.FireBolt, id, 4, step.Distance(1, 5))
			} else {
				s.logger.Debug("No secondary skills available, using primary attack")
				step.PrimaryAttack(id, 1, false, step.Distance(1, 3))
			}
		}

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s SorceressLevelingLightning) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s SorceressLevelingLightning) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.ThunderStorm); found {
		skillsList = append(skillsList, skill.ThunderStorm)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			break
		}
	}

	return skillsList
}

func (s SorceressLevelingLightning) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s SorceressLevelingLightning) staticFieldCasts() int {
	casts := 6
	ctx := context.Get()

	switch ctx.CharacterCfg.Game.Difficulty {
	case difficulty.Normal:
		casts = 8
	}
	s.logger.Debug("Static Field casts", "count", casts)
	return casts
}

func (s SorceressLevelingLightning) ShouldResetSkills() bool {
	lvl, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 25 && s.data.PlayerUnit.Skills[skill.Nova].Level > 10 {
		s.logger.Info("Resetting skills: Level 25+ and Nova level > 10")
		return true
	}

	return false
}

func (s SorceressLevelingLightning) SkillsToBind() (skill.ID, []skill.ID) {
	level, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
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

	if s.data.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		skillBindings = append(skillBindings, skill.Blizzard)
	} else if s.data.PlayerUnit.Skills[skill.Nova].Level > 1 {
		skillBindings = append(skillBindings, skill.Nova)
	} else if s.data.PlayerUnit.Skills[skill.ChargedBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.ChargedBolt)
	} else if s.data.PlayerUnit.Skills[skill.FireBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.FireBolt)
	}

	mainSkill := skill.AttackSkill
	if s.data.PlayerUnit.Skills[skill.GlacialSpike].Level > 0 {
		mainSkill = skill.GlacialSpike
	}

	s.logger.Info("Skills bound", "mainSkill", mainSkill, "skillBindings", skillBindings)
	return mainSkill, skillBindings
}

func (s SorceressLevelingLightning) StatPoints() map[stat.ID]int {
	lvl, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
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

func (s SorceressLevelingLightning) SkillPoints() []skill.ID {
	lvl, _ := s.data.PlayerUnit.FindStat(stat.Level, 0)
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

func (s SorceressLevelingLightning) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingLightning) KillAndariel() error {
	m, _ := s.data.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(3, 5))

	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s SorceressLevelingLightning) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s SorceressLevelingLightning) KillDuriel() error {
	m, _ := s.data.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s SorceressLevelingLightning) KillCouncil() error {
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

func (s SorceressLevelingLightning) KillMephisto() error {
	m, _ := s.data.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s SorceressLevelingLightning) KillIzual() error {
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

func (s SorceressLevelingLightning) KillDiablo() error {
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

func (s SorceressLevelingLightning) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingLightning) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLevelingLightning) KillAncients() error {
	for _, m := range s.data.Monsters.Enemies(data.MonsterEliteFilter()) {
		m, _ := s.data.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)

		step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(8, 10))

		step.MoveTo(data.Position{X: 10062, Y: 12639})

		s.killMonster(m.Name, data.MonsterTypeSuperUnique)
	}
	return nil
}

func (s SorceressLevelingLightning) KillBaal() error {
	m, _ := s.data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
	step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 4))
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)

	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
