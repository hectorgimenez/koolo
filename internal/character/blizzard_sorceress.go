package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Blizzard, skill.Teleport, skill.TomeOfTownPortal, skill.ShiverArmor}

	if s.Data.CharacterCfg.Character.Sorceress.UseStaticField {
		requireKeybindings = append(requireKeybindings, skill.StaticField)
	}

	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.ShiverArmor:
				_, found1 := s.Data.KeyBindings.KeyBindingForSkill(skill.FrozenArmor)
				_, found2 := s.Data.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
				if !found1 && !found2 {
					missingKeybindings = append(missingKeybindings, skill.ShiverArmor)
				}
			default:
				missingKeybindings = append(missingKeybindings, cskill)
			}
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s BlizzardSorceress) killMonsterWithStatic(bossID npc.ID, monsterType data.MonsterType) error {
	for {
		boss, found := s.Data.Monsters.FindOne(bossID, monsterType)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil
		}

		bossHPPercent := (float64(boss.Stats[stat.Life]) / float64(boss.Stats[stat.MaxLife])) * 100
		thresholdFloat := float64(s.Data.CharacterCfg.Character.Sorceress.BossStaticThreshold)

		// Cast Static Field until boss HP is below threshold
		if bossHPPercent > thresholdFloat {
			staticOpts := []step.AttackOption{
				step.Distance(s.Data.CharacterCfg.Character.Sorceress.StaticFieldMinDist, s.Data.CharacterCfg.Character.Sorceress.StaticFieldMaxDist),
			}
			err := step.SecondaryAttack(skill.StaticField, boss.UnitID, 1, staticOpts...)
			if err != nil {
				s.Logger.Warn("Failed to cast Static Field", slog.String("error", err.Error()))
			}
			continue
		}

		// Switch to primary skill once boss HP is low enough
		return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			return boss.UnitID, true
		}, nil)
	}
}

func (s BlizzardSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	previousSelfBlizzard := time.Time{}

	blizzOpts := step.StationaryDistance(
		s.Data.CharacterCfg.Character.Sorceress.RightSkillMinDist,
		s.Data.CharacterCfg.Character.Sorceress.RightSkillMaxDist,
	)
	lsOpts := step.Distance(
		s.Data.CharacterCfg.Character.Sorceress.LeftSkillMinDist,
		s.Data.CharacterCfg.Character.Sorceress.LeftSkillMinDist,
	)

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

		if completedAttackLoops >= s.Data.CharacterCfg.Character.Sorceress.MaxAttacksLoop {
			s.Logger.Error("Exceeded MaxAttacksLoop", slog.String("completedAttackLoops", fmt.Sprintf("%v", completedAttackLoops)))
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}
		if monster.Stats[stat.Life] <= 0 {
			return nil
		}

		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		if time.Since(previousSelfBlizzard) > time.Second*4 && !s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			for _, m := range s.Data.Monsters.Enemies() {
				if dist := s.PathFinder.DistanceFromMe(m.Position); dist < 4 {
					previousSelfBlizzard = time.Now()
					step.SecondaryAttack(skill.Blizzard, m.UnitID, 1, blizzOpts)
				}
			}
		}

		for s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			step.PrimaryAttack(id, 2, true, lsOpts)
			time.Sleep(50 * time.Millisecond) // Give time for the attack to complete
		}

		step.SecondaryAttack(skill.Blizzard, id, 1, blizzOpts)

		completedAttackLoops++

		previousUnitID = int(id)
	}
}

func (s BlizzardSorceress) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s BlizzardSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s BlizzardSorceress) BuffSkills() []skill.ID {
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

func (s BlizzardSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s BlizzardSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s BlizzardSorceress) KillAndariel() error {
	if s.Data.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Andariel, data.MonsterTypeUnique)
	}

	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, nil)
}

func (s BlizzardSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s BlizzardSorceress) KillDuriel() error {
	if s.Data.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Duriel, data.MonsterTypeUnique)
	}

	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
}

func (s BlizzardSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
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
	}, nil)
}

func (s BlizzardSorceress) KillMephisto() error {
	if s.Data.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Mephisto, data.MonsterTypeUnique)
	}

	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}

func (s BlizzardSorceress) KillIzual() error {
	if s.Data.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Izual, data.MonsterTypeUnique)
	}

	return s.killMonsterByName(npc.Izual, data.MonsterTypeUnique, nil)
}

func (s BlizzardSorceress) KillDiablo() error {
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

		if s.Data.CharacterCfg.Character.Sorceress.UseStaticField {
			return s.killMonsterWithStatic(npc.Diablo, data.MonsterTypeUnique)
		}

		return s.killMonsterByName(npc.Diablo, data.MonsterTypeUnique, nil)
	}
}

func (s BlizzardSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s BlizzardSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (s BlizzardSorceress) KillBaal() error {
	if s.Data.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.BaalCrab, data.MonsterTypeUnique)
	}

	return s.killMonsterByName(npc.BaalCrab, data.MonsterTypeUnique, nil)
}
