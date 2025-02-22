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

const (
	orbMaxAttacksLoop = 40
	orbMinDistance    = 8
	orbMaxDistance    = 20
	orbLSMinDistance  = 6
	orbLSMaxDistance  = 15 // Left skill
)

type FrozenOrbSorceress struct {
	BaseCharacter
}

func (s FrozenOrbSorceress) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.FrozenOrb, skill.Teleport, skill.TomeOfTownPortal, skill.ShiverArmor, skill.StaticField}
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

func (s FrozenOrbSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	previousSelfOrb := time.Time{}

	orbOpts := step.StationaryDistance(orbMinDistance, orbMaxDistance)
	lsOpts := step.Distance(orbLSMinDistance, orbLSMaxDistance)

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

		if completedAttackLoops >= orbMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		if time.Since(previousSelfOrb) > time.Second*4 && !s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			for _, m := range s.Data.Monsters.Enemies() {
				if dist := s.PathFinder.DistanceFromMe(m.Position); dist < 4 {
					previousSelfOrb = time.Now()
					step.SecondaryAttack(skill.FrozenOrb, m.UnitID, 1, orbOpts)
				}
			}
		}

		if s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			step.PrimaryAttack(id, 2, true, lsOpts)
		}

		step.SecondaryAttack(skill.FrozenOrb, id, 1, orbOpts)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s FrozenOrbSorceress) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s FrozenOrbSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s FrozenOrbSorceress) BuffSkills() []skill.ID {
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

func (s FrozenOrbSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s FrozenOrbSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s FrozenOrbSorceress) KillAndariel() error {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, nil)
}

func (s FrozenOrbSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s FrozenOrbSorceress) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
}

func (s FrozenOrbSorceress) KillCouncil() error {
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

func (s FrozenOrbSorceress) KillMephisto() error {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}

func (s FrozenOrbSorceress) KillIzual() error {
	m, _ := s.Data.Monsters.FindOne(npc.Izual, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))

	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s FrozenOrbSorceress) KillDiablo() error {
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

		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s FrozenOrbSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s FrozenOrbSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (s FrozenOrbSorceress) KillBaal() error {
	m, _ := s.Data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeUnique)
	step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))

	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
