package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	hydra_sorceressMaxAttacksLoop = 40
	hydra_sorceressMinDistance    = 15
	hydra_sorceressMaxDistance    = 30
	hydra_SFMinDistance           = 4
	hydra_SFMaxDistance           = 6
)

type HydraSorceress struct {
	BaseCharacter
}

func (s HydraSorceress) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Hydra, skill.Teleport, skill.TomeOfTownPortal, skill.ShiverArmor, skill.StaticField}
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

func (s HydraSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	skipOnImmunities = append(skipOnImmunities, stat.FireImmune)
	opts := step.Distance(hydra_sorceressMinDistance, hydra_sorceressMaxDistance)

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

		if completedAttackLoops >= hydra_sorceressMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		step.SecondaryAttack(skill.Hydra, id, 6, opts)
		step.PrimaryAttack(id, 3, false, opts)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s HydraSorceress) KillBossSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {

	opts := step.Distance(hydra_sorceressMinDistance, hydra_sorceressMaxDistance)
	sfOpts := step.Distance(hydra_SFMinDistance, hydra_SFMaxDistance)

	id, found := monsterSelector(*s.Data)
	if !found {
		return nil
	}

	monster, found := s.Data.Monsters.FindByID(id)
	if !found {
		s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
		return nil
	}

	_ = step.SecondaryAttack(skill.StaticField, monster.UnitID, 6, sfOpts)

	for monster.Stats[stat.Life] > 0 && found {
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}
		step.SecondaryAttack(skill.Hydra, id, 6, opts)
		step.PrimaryAttack(id, 6, false, opts)
	}

	return nil
}

func (s HydraSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, _ int, _ bool, skipOnImmunities []stat.Resist) error {
	return s.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s HydraSorceress) BuffSkills() []skill.ID {
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

func (s HydraSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s HydraSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, hydra_sorceressMaxDistance, false, nil)
}

func (s HydraSorceress) KillAndariel() error {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, hydra_sorceressMaxDistance, false, nil)
}
func (s HydraSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, hydra_sorceressMaxDistance, false, nil)
}

func (s HydraSorceress) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, hydra_sorceressMaxDistance, true, nil)
}

func (s HydraSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var veryImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.ColdImmune) && m.IsImmune(stat.FireImmune) {
					veryImmunes = append(veryImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}
		}

		councilMembers = append(councilMembers, veryImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s HydraSorceress) KillMephisto() error {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, hydra_sorceressMaxDistance, true, nil)
}

func (s HydraSorceress) KillIzual() error {
	m, _ := s.Data.Monsters.FindOne(npc.Izual, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))

	return s.killMonsterByName(npc.Izual, data.MonsterTypeUnique, hydra_sorceressMaxDistance, true, nil)
}

func (s HydraSorceress) KillDiablo() error {
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

		_ = step.SecondaryAttack(skill.StaticField, diablo.UnitID, 5, step.Distance(3, 8))

		return s.killMonsterByName(npc.Diablo, data.MonsterTypeUnique, hydra_sorceressMaxDistance, true, nil)
	}
}

func (s HydraSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, hydra_sorceressMaxDistance, false, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s HydraSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, hydra_sorceressMaxDistance, false, nil)
}

func (s HydraSorceress) KillBaal() error {
	m, _ := s.Data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeUnique)
	step.SecondaryAttack(skill.StaticField, m.UnitID, 5, step.Distance(5, 8))

	return s.killMonsterByName(npc.BaalCrab, data.MonsterTypeUnique, hydra_sorceressMaxDistance, true, nil)
}
