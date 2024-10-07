package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

const (
	NovaSorceressMaxAttacksLoop = 10
	NovaSorceressMinDistance    = 8
	NovaSorceressMaxDistance    = 13
)

type NovaSorceress struct {
	BaseCharacter
}

func (s NovaSorceress) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Nova, skill.Teleport, skill.TomeOfTownPortal, skill.FrozenArmor, skill.StaticField}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.FrozenArmor:
				_, found1 := s.data.KeyBindings.KeyBindingForSkill(skill.ShiverArmor)
				_, found2 := s.data.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
				if !found1 && !found2 {
					missingKeybindings = append(missingKeybindings, skill.FrozenArmor)
				}
			default:
				missingKeybindings = append(missingKeybindings, cskill)
			}
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s NovaSorceress) KillMonsterSequence(
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

		opts := step.Distance(NovaSorceressMinDistance, NovaSorceressMaxDistance)

		if s.shouldCastStatic() {
			step.SecondaryAttack(skill.StaticField, id, 1, opts)
		}

		// In case monster is stuck behind a wall or character is not able to reach it we will short the distance
		if completedAttackLoops > 5 {
			if completedAttackLoops == 6 {
				s.logger.Debug("Looks like monster is not reachable, reducing max attack distance.")
			}
			opts = step.Distance(0, 1)
		}

		step.SecondaryAttack(skill.Nova, id, 5, opts)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s NovaSorceress) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s NovaSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, maxDistance int, _ bool, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s NovaSorceress) shouldCastStatic() bool {

	// Iterate through all mobs within max range and collect them
	mobs := make([]data.Monster, 0)

	for _, m := range s.data.Monsters.Enemies() {
		if s.pf.DistanceFromMe(m.Position) <= NovaSorceressMaxDistance+5 {
			mobs = append(mobs, m)
		} else {
			continue
		}
	}

	// Iterate through the mob list and check their if more than 50% of the mobs are above 60% hp
	var mobsAbove60Percent int
	for _, mob := range mobs {

		life := mob.Stats[stat.Life]
		maxLife := mob.Stats[stat.MaxLife]

		lifePercentage := int((float64(life) / float64(maxLife)) * 100)

		if lifePercentage > 60 {
			mobsAbove60Percent++
		}
	}

	return mobsAbove60Percent > len(mobs)/2
}

func (s NovaSorceress) BuffSkills() []skill.ID {
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
			return skillsList
		}
	}

	return skillsList
}

func (s NovaSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s NovaSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, NovaSorceressMaxDistance, false, nil)
}

func (s NovaSorceress) KillAndariel() error {
	m, _ := s.data.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 7, step.Distance(8, 13))

	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s NovaSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeNone, NovaSorceressMaxDistance, false, nil)
}

func (s NovaSorceress) KillDuriel() error {
	m, _ := s.data.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 7, step.Distance(8, 13))

	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s NovaSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var lightningImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.LightImmune) {
					lightningImmunes = append(lightningImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}
		}

		councilMembers = append(councilMembers, lightningImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s NovaSorceress) KillMephisto() error {
	m, _ := s.data.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 7, step.Distance(8, 13))

	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s NovaSorceress) KillIzual() error {
	m, _ := s.data.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 7, step.Distance(8, 13))
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)

	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s NovaSorceress) KillDiablo() error {
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

		_ = step.SecondaryAttack(skill.StaticField, diablo.UnitID, 5, step.Distance(8, 13))

		return s.killMonster(npc.Diablo, data.MonsterTypeNone)
	}
}

func (s NovaSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, s.cfg.Game.Pindleskin.SkipOnImmunities)
}

func (s NovaSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, NovaSorceressMaxDistance, false, nil)
}

func (s NovaSorceress) KillBaal() error {
	m, _ := s.data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
	step.SecondaryAttack(skill.StaticField, m.UnitID, 5, step.Distance(8, 13))
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)

	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
