package character

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	sorceressMaxAttacksLoop = 40
	sorceressMinDistance    = 25
	sorceressMaxDistance    = 30
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.Blizzard, skill.Teleport, skill.TomeOfTownPortal, skill.ShiverArmor, skill.StaticField}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := d.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.ShiverArmor:
				_, found1 := d.KeyBindings.KeyBindingForSkill(skill.FrozenArmor)
				_, found2 := d.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
				if !found1 && !found2 {
					missingKeybindings = append(missingKeybindings, skill.ShiverArmor)
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

func (s BlizzardSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0
	previousSelfBlizzard := time.Time{}

	return action.NewStepChain(func(d game.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		if len(opts) == 0 {
			opts = append(opts, step.Distance(sorceressMinDistance, sorceressMaxDistance))
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}
		}

		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		if time.Since(previousSelfBlizzard) > time.Second*4 && !d.PlayerUnit.States.HasState(state.Cooldown) {
			for _, m := range d.Monsters.Enemies() {
				if dist := pather.DistanceFromMe(d, m.Position); dist < 4 {
					s.logger.Debug("Monster detected close to the player, casting Blizzard over it")
					previousSelfBlizzard = time.Now()
					return []step.Step{step.SecondaryAttack(skill.Blizzard, m.UnitID, 1, opts...)}
				}
			}
		}

		// In case monster is stuck behind a wall or character is not able to reach it we will short the distance
		if completedAttackLoops > 12 {
			if completedAttackLoops == 13 {
				s.logger.Debug("Looks like monster is not reachable, reducing max attack distance")
			}
			opts = []step.AttackOption{step.Distance(1, 5)}
		}

		completedAttackLoops++
		previousUnitID = int(id)

		if d.PlayerUnit.States.HasState(state.Cooldown) {
			return []step.Step{step.PrimaryAttack(id, 2, true, opts...)}
		}

		return []step.Step{step.SecondaryAttack(skill.Blizzard, id, 1, opts...)}
	}, action.RepeatUntilNoSteps())
}

func (s BlizzardSorceress) BuffSkills(d game.Data) []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := d.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			return skillsList
		}
	}

	return skillsList
}

func (s BlizzardSorceress) PreCTABuffSkills(d game.Data) []skill.ID {
	return []skill.ID{}
}

func (s BlizzardSorceress) KillCountess() action.Action {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillAndariel() action.Action {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillSummoner() action.Action {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillDuriel() action.Action {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeNone, sorceressMaxDistance, true, nil)
}

func (s BlizzardSorceress) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, skipOnImmunities)
}
func (s BlizzardSorceress) KillMephisto() action.Action {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeNone, sorceressMaxDistance, true, nil)
}

func (s BlizzardSorceress) KillNihlathak() action.Action {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			return []action.Action{action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			})}
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, diablo.UnitID, 5, step.Distance(3, 8)),
				}
			}),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s BlizzardSorceress) KillIzual() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, 7, step.Distance(5, 8)),
				}
			}),
			// We will need a lot of cycles to kill him probably
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

func (s BlizzardSorceress) KillBaal() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, 5, step.Distance(5, 8)),
				}
			}),
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
		}
	})
}

func (s BlizzardSorceress) KillCouncil() action.Action {
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
	}, nil, step.Distance(8, sorceressMaxDistance))
}

func (s BlizzardSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, maxDistance int, _ bool, skipOnImmunities []stat.Resist) action.Action {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities, step.Distance(sorceressMinDistance, maxDistance))
}

func (s BlizzardSorceress) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}
