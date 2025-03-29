package character

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
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

const (
	SorceressLevelingMaxAttacksLoop = 35
	SorceressLevelingMinDistance    = 8
	SorceressLevelingMaxDistance    = 20
	SorceressLevelingMeleeDistance  = 2
)

func (s SorceressLeveling) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
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
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= SorceressLevelingMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		attackSuccess := false
		lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
		if s.Data.PlayerUnit.MPPercent() < 15 && lvl.Value < 15 {
			s.Logger.Debug("Low mana, using primary attack")
			step.PrimaryAttack(id, 1, false, step.Distance(1, SorceressLevelingMeleeDistance))
		} else {
			if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Blizzard); found {
				if s.Data.PlayerUnit.Mode == mode.CastingSkill {
					attackSuccess = true
					s.Logger.Debug("Using Blizzard")
				}
				step.SecondaryAttack(skill.Blizzard, id, 1, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Meteor); found {
				if s.Data.PlayerUnit.Mode == mode.CastingSkill {
					attackSuccess = true
					s.Logger.Debug("Using Meteor")
				}
				step.SecondaryAttack(skill.Meteor, id, 1, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.FireBall); found {
				if s.Data.PlayerUnit.Mode == mode.CastingSkill {
					attackSuccess = true
					s.Logger.Debug("Using FireBall")
				}
				step.SecondaryAttack(skill.FireBall, id, 1, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.FireBolt); found {
				if s.Data.PlayerUnit.Mode == mode.CastingSkill {
					attackSuccess = true
					s.Logger.Debug("Using FireBolt")
				}
				step.SecondaryAttack(skill.FireBolt, id, 1, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else {
				s.Logger.Debug("No secondary skills available, using primary attack")
				step.PrimaryAttack(id, 1, false, step.Distance(1, SorceressLevelingMeleeDistance))
			}
		}
		if attackSuccess {
			completedAttackLoops++
		}
		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s SorceressLeveling) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}
func (s SorceressLeveling) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			return skillsList
		}
	}

	return skillsList
}

func (s SorceressLeveling) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s SorceressLeveling) ShouldResetSkills() bool {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 32 && s.Data.PlayerUnit.Skills[skill.FireBall].Level > 1 {
		s.Logger.Info("Respecing to Blizzard: Level 32+ and FireBall level > 1")
		return true
	}
	return false
}

func (s SorceressLeveling) SkillsToBind() (skill.ID, []skill.ID) {
	level, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	skillBindings := []skill.ID{}

	if level.Value >= 4 {
		skillBindings = append(skillBindings, skill.FrozenArmor)
	}
	if level.Value >= 6 {
		skillBindings = append(skillBindings, skill.StaticField)
	}
	if level.Value >= 18 {
		skillBindings = append(skillBindings, skill.Teleport)
	}

	if s.Data.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		skillBindings = append(skillBindings, skill.Blizzard)
	} else if s.Data.PlayerUnit.Skills[skill.Meteor].Level > 0 {
		skillBindings = append(skillBindings, skill.Meteor)
	} else if s.Data.PlayerUnit.Skills[skill.Hydra].Level > 0 {
		skillBindings = append(skillBindings, skill.Hydra)
	} else if s.Data.PlayerUnit.Skills[skill.FireBall].Level > 0 {
		skillBindings = append(skillBindings, skill.FireBall)
	} else if s.Data.PlayerUnit.Skills[skill.FireBolt].Level > 0 {
		skillBindings = append(skillBindings, skill.FireBolt)
	}

	mainSkill := skill.AttackSkill
	if s.Data.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		mainSkill = skill.Blizzard
	} else if s.Data.PlayerUnit.Skills[skill.Meteor].Level > 0 {
		mainSkill = skill.Meteor
	}

	s.Logger.Info("Skills bound", "mainSkill", mainSkill, "skillBindings", skillBindings)
	return mainSkill, skillBindings
}

func (s SorceressLeveling) StatPoints() []context.StatAllocation {

	// Define target totals (including base stats)
	targets := []context.StatAllocation{
		{Stat: stat.Vitality, Points: 50},  // Base 10 + 40
		{Stat: stat.Strength, Points: 25},  // Base 10 + 15
		{Stat: stat.Vitality, Points: 65},  // Previous 50 + 15
		{Stat: stat.Strength, Points: 47},  // Previous 25 + 22
		{Stat: stat.Vitality, Points: 999}, // Rest into vit
	}

	return targets
}

func (s SorceressLeveling) SkillPoints() []skill.ID {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	var skillPoints []skill.ID

	if lvl.Value < 32 {
		skillPoints = []skill.ID{
			skill.FireBolt,    // 2
			skill.FireBolt,    // 3
			skill.FireBolt,    // 4
			skill.FrozenArmor, // 5
			skill.StaticField, // 6
			skill.Warmth,      // 7
			skill.FireBolt,    // 8
			skill.FireBolt,    // 9
			skill.FireBolt,    // 10
			skill.FireBolt,    // 11
			skill.FireBolt,    // 12
			skill.FireBall,    // 13
			skill.FireBall,    // 14
			skill.FireBall,    // 15
			skill.FireBall,    // 16
			skill.Telekinesis, // 17
			skill.Teleport,    // 18
			skill.FireBall,    // 19
			skill.FireBall,    // 20
			skill.FireBall,    // 21
			skill.FireBall,    // 22
			skill.FireBall,    // 23
			skill.FireBall,    // 24
			skill.FireBall,    // 25
			skill.FireBall,    // 26
			skill.FireBall,    // 27
			skill.FireBall,    // 28
			skill.FireBall,    // 29
			skill.FireMastery, // 30
			skill.FireBall,    // 31
			skill.FireBall,    // 32
			skill.FireBall,    // 33
			skill.FireBall,    // 34
			skill.FireBall,    // 35
			skill.FireBolt,    // 36 Let's overshoot by 4 in case we got the 4 skill quests
		}
	} else {
		skillPoints = []skill.ID{
			skill.StaticField,
			skill.Telekinesis,
			skill.Teleport,
			skill.FrozenArmor,
			skill.IceBolt,
			skill.IceBlast,
			skill.GlacialSpike,
			skill.FrostNova,
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
			skill.GlacialSpike,
			skill.GlacialSpike,
			skill.GlacialSpike,
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
			skill.ColdMastery,
			skill.ColdMastery,
			skill.ColdMastery,
			skill.ColdMastery,
			skill.ColdMastery,
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
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBlast,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
			skill.IceBolt,
		}
	}

	s.Logger.Info("Assigning skill points", "level", lvl.Value, "skillPoints", skillPoints)
	return skillPoints
}

func (s SorceressLeveling) KillCountess() error {
	m, _ := s.Data.Monsters.FindOne(npc.DarkStalker, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s SorceressLeveling) KillAndariel() error {
	m, _ := s.Data.Monsters.FindOne(npc.Andariel, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, nil)
}
func (s SorceressLeveling) KillSummoner() error {
	m, _ := s.Data.Monsters.FindOne(npc.Summoner, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s SorceressLeveling) KillDuriel() error {
	m, _ := s.Data.Monsters.FindOne(npc.Duriel, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
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
			distanceI := s.PathFinder.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.PathFinder.DistanceFromMe(councilMembers[j].Position)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s SorceressLeveling) KillMephisto() error {
	m, _ := s.Data.Monsters.FindOne(npc.Mephisto, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}
func (s SorceressLeveling) KillIzual() error {
	m, _ := s.Data.Monsters.FindOne(npc.Izual, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))
	return s.killMonsterByName(npc.Izual, data.MonsterTypeUnique, nil)
}

func (s SorceressLeveling) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
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
		s.Logger.Info("Diablo detected, attacking")

		_ = step.SecondaryAttack(skill.StaticField, diablo.UnitID, 5, step.Distance(1, 2))

		return s.killMonsterByName(npc.Diablo, data.MonsterTypeUnique, nil)
	}
}

func (s SorceressLeveling) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s SorceressLeveling) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (s SorceressLeveling) KillAncients() error {
	for _, m := range s.Data.Monsters.Enemies(data.MonsterEliteFilter()) {
		m, _ := s.Data.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)
		step.MoveTo(data.Position{X: 10062, Y: 12639})
		_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))
		s.killMonsterByName(m.Name, data.MonsterTypeSuperUnique, nil)
	}
	return nil
}

func (s SorceressLeveling) KillBaal() error {
	m, _ := s.Data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeUnique)
	step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(1, 2))

	return s.killMonsterByName(npc.BaalCrab, data.MonsterTypeUnique, nil)
}
