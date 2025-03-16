package character

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	hydraSorceressMaxAttacksLoop = 40
	hydraSorceressMinDistance    = 7
	hydraSorceressMaxDistance    = 15
)

type HydraSorceress struct {
	BaseCharacter
}

func (f HydraSorceress) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Hydra, skill.Teleport, skill.TomeOfTownPortal, skill.FrozenArmor, skill.StaticField}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := f.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.FrozenArmor:

				_, found1 := f.Data.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
				_, found2 := f.Data.KeyBindings.KeyBindingForSkill(skill.ShiverArmor)
				if !found1 && !found2 {
					missingKeybindings = append(missingKeybindings, skill.FrozenArmor)
				}

			default:
				missingKeybindings = append(missingKeybindings, cskill)
			}
		}
	}

	if len(missingKeybindings) > 0 {
		f.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (f HydraSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	canCastHydra := true

	// Setup timer for Hydra casting
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			canCastHydra = true
		}
	}()

	distanceOrgs := step.RangedDistance(0, (hydraSorceressMaxDistance + 5))

	for {
		id, found := monsterSelector(*f.Data)
		if !found {
			return nil
		}

		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !f.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= hydraSorceressMaxAttacksLoop {
			return nil
		}

		monster, found := f.Data.Monsters.FindByID(id)
		if !found {
			f.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// Kite Monsters
		for _, m := range f.Data.Monsters.Enemies() {
			if dist := f.PathFinder.DistanceFromMe(m.Position); dist < hydraSorceressMinDistance {
				step.MoveTo(kitePosition(m.Position))
				step.SecondaryAttack(skill.Hydra, id, 1, distanceOrgs)
				continue
			}
		}

		// Hydra or Fireball
		if canCastHydra {
			step.SecondaryAttack(skill.Hydra, id, 3, distanceOrgs)
			canCastHydra = false
		} else {
			step.PrimaryAttack(id, 2, true, distanceOrgs)
		}

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func kitePosition(monsterPos data.Position) data.Position {
	xOffset := rand.Intn(hydraSorceressMaxDistance-hydraSorceressMinDistance+1) + hydraSorceressMinDistance
	yOffset := rand.Intn(hydraSorceressMaxDistance-hydraSorceressMinDistance+1) + hydraSorceressMinDistance

	if rand.Float32() < 0.5 {
		xOffset = -xOffset
	}
	if rand.Float32() < 0.5 {
		yOffset = -yOffset
	}

	return data.Position{
		X: monsterPos.X + xOffset,
		Y: monsterPos.Y + yOffset,
	}
}

func (f HydraSorceress) killMonster(npc npc.ID, t data.MonsterType) error {
	return f.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {

			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (f HydraSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return f.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {

			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (f HydraSorceress) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := f.Data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := f.Data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			return skillsList
		}
	}

	return skillsList
}

func (f HydraSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (f HydraSorceress) KillCountess() error {
	return f.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (f HydraSorceress) KillAndariel() error {
	return f.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, nil)
}

func (f HydraSorceress) KillSummoner() error {
	return f.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (f HydraSorceress) KillDuriel() error {
	return f.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
}

func (f HydraSorceress) KillCouncil() error {
	return f.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var fireImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {

			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.FireImmune) {
					fireImmunes = append(fireImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}

		}

		councilMembers = append(councilMembers, fireImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (f HydraSorceress) KillMephisto() error {
	return f.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}

func (f HydraSorceress) KillIzual() error {
	m, _ := f.Data.Monsters.FindOne(npc.Izual, data.MonsterTypeUnique)
	_ = step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))

	return f.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (f HydraSorceress) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()

	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			f.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := f.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)

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
		f.Logger.Info("Diablo detected, attacking")

		_ = step.SecondaryAttack(skill.StaticField, diablo.UnitID, 5, step.Distance(3, 8))

		return f.killMonster(npc.Diablo, data.MonsterTypeUnique)

	}
}

func (f HydraSorceress) KillPindle() error {
	return f.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, f.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (f HydraSorceress) KillNihlathak() error {
	return f.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (f HydraSorceress) KillBaal() error {
	m, _ := f.Data.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeUnique)
	step.SecondaryAttack(skill.StaticField, m.UnitID, 4, step.Distance(5, 8))

	return f.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
