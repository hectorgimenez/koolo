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

const (
	SorceressLevelingMaxAttacksLoop = 10
	SorceressLevelingMinDistance    = 25
	SorceressLevelingMaxDistance    = 30
	SorceressLevelingMeleeDistance  = 3
)

func (s SorceressLeveling) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.TomeOfTownPortal}
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

		lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
		if s.Data.PlayerUnit.MPPercent() < 15 && lvl.Value < 15 {
			s.Logger.Debug("Low mana, using primary attack")
			step.PrimaryAttack(id, 1, false, step.Distance(1, SorceressLevelingMeleeDistance))
		} else {
			if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Blizzard); found {
				s.Logger.Debug("Using Blizzard")
				step.SecondaryAttack(skill.Blizzard, id, 1, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Meteor); found {
				s.Logger.Debug("Using Meteor")
				step.SecondaryAttack(skill.Meteor, id, 1, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.FireBall); found {
				s.Logger.Debug("Using FireBall")
				step.SecondaryAttack(skill.FireBall, id, 4, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.FireBolt); found {
				s.Logger.Debug("Using FireBolt")
				step.SecondaryAttack(skill.FireBolt, id, 4, step.Distance(SorceressLevelingMinDistance, SorceressLevelingMaxDistance))
			} else {
				s.Logger.Debug("No secondary skills available, using primary attack")
				step.PrimaryAttack(id, 1, false, step.Distance(1, SorceressLevelingMeleeDistance))
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
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.FrozenArmor); found {
		skillsList = append(skillsList, skill.FrozenArmor)
	}

	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	s.Logger.Info("Buff skills", "skills", skillsList)
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
	s.Logger.Debug("Static Field casts", "count", casts)
	return casts
}

func (s SorceressLeveling) ShouldResetSkills() bool {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 24 && s.Data.PlayerUnit.Skills[skill.FireBall].Level > 1 {
		s.Logger.Info("Resetting skills: Level 24+ and FireBall level > 1")
		return true
	}
	return false
}

func (s SorceressLeveling) SkillsToBind() (skill.ID, []skill.ID) {
	level, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
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

	if s.Data.PlayerUnit.Skills[skill.Blizzard].Level > 0 {
		skillBindings = append(skillBindings, skill.Blizzard)
	} else if s.Data.PlayerUnit.Skills[skill.Meteor].Level > 0 {
		skillBindings = append(skillBindings, skill.Meteor)
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

func (s SorceressLeveling) StatPoints() map[stat.ID]int {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	statPoints := make(map[stat.ID]int)

	if lvl.Value < 9 {
		statPoints[stat.Strength] = 15
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

	s.Logger.Info("Assigning stat points", "level", lvl.Value, "statPoints", statPoints)
	return statPoints
}

func (s SorceressLeveling) SkillPoints() []skill.ID {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
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
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.Teleport,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
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

	s.Logger.Info("Assigning skill points", "level", lvl.Value, "skillPoints", skillPoints)
	return skillPoints
}

func (s SorceressLeveling) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeUnique)
}
func (s SorceressLeveling) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeUnique)
}

func (s SorceressLeveling) KillDuriel() error {
	m, _ := s.Data.Monsters.FindOne(npc.Duriel, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

	return s.killMonster(npc.Duriel, data.MonsterTypeUnique)
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
	return s.killMonster(npc.Mephisto, data.MonsterTypeUnique)
}
func (s SorceressLeveling) KillIzual() error {
	m, _ := s.Data.Monsters.FindOne(npc.Izual, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
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

		_ = step.SecondaryAttack(skill.StaticField, diablo.UnitID, s.staticFieldCasts(), step.Distance(1, 5))

		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s SorceressLeveling) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAncients() error {
	for _, m := range s.Data.Monsters.Enemies(data.MonsterEliteFilter()) {
		m, _ := s.Data.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)

		step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(8, 10))

		step.MoveTo(data.Position{X: 10062, Y: 12639})

		s.killMonster(m.Name, data.MonsterTypeSuperUnique)
	}
	return nil
}

func (s SorceressLeveling) KillBaal() error {
	m, _ := s.Data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeUnique)
	step.SecondaryAttack(skill.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 4))

	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
