package character

import (
	"log/slog"
	"time"

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
	lightningSorceressMaxAttacksLoop = 10
	lightningSorceressMinDistance    = 8
	lightningSorceressMaxDistance    = 13
)

type LightningSorceress struct {
	BaseCharacter
}

func (s LightningSorceress) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.Nova, skill.Teleport, skill.TomeOfTownPortal, skill.FrozenArmor, skill.StaticField}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := d.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.FrozenArmor:
				_, found1 := d.KeyBindings.KeyBindingForSkill(skill.ShiverArmor)
				_, found2 := d.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
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

func (s LightningSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0

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
			opts = append(opts, step.Distance(lightningSorceressMinDistance, lightningSorceressMaxDistance))
		}

		if completedAttackLoops >= lightningSorceressMaxAttacksLoop {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		// Check if we should use static
		if s.shouldCastStatic(d) {
			steps = append(steps, step.SecondaryAttack(skill.StaticField, id, 1, opts...))
		}

		if completedAttackLoops%2 == 0 {
			for _, m := range d.Monsters.Enemies() {
				if d := pather.DistanceFromMe(d, m.Position); d < 5 {
					s.logger.Debug("Monster detected close to the player, casting Nova over it")
					steps = append(steps, step.SecondaryAttack(skill.Nova, m.UnitID, 1, opts...))
					break
				}
			}
		}

		// In case monster is stuck behind a wall or character is not able to reach it we will short the distance
		if completedAttackLoops > 5 {
			if completedAttackLoops == 6 {
				s.logger.Debug("Looks like monster is not reachable, reducing max attack distance.")
			}
			opts = []step.AttackOption{step.Distance(0, 1)}
		}

		steps = append(steps,
			step.SecondaryAttack(skill.Nova, id, 5, opts...),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s LightningSorceress) BuffSkills(d game.Data) []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	if _, found := d.KeyBindings.KeyBindingForSkill(skill.ThunderStorm); found {
		skillsList = append(skillsList, skill.ThunderStorm)
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

func (s LightningSorceress) PreCTABuffSkills(d game.Data) []skill.ID {
	return []skill.ID{}
}

func (s LightningSorceress) shouldCastStatic(d game.Data) bool {

	// Iterate through all mobs within max range and collect them
	mobs := make([]data.Monster, 0)

	for _, m := range d.Monsters.Enemies() {
		if pather.DistanceFromMe(d, m.Position) <= lightningSorceressMaxDistance+5 {
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

func (s LightningSorceress) KillCountess() action.Action {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, lightningSorceressMaxDistance, false, nil)
}

func (s LightningSorceress) KillAndariel() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, 3, step.Distance(lightningSorceressMaxDistance, 15)),
				}
			}),
			s.killMonsterByName(npc.Andariel, data.MonsterTypeNone, lightningSorceressMaxDistance, false, nil),
		}
	})
}

func (s LightningSorceress) KillSummoner() action.Action {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeNone, lightningSorceressMaxDistance, false, nil)
}

func (s LightningSorceress) KillDuriel() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, 3, step.Distance(lightningSorceressMaxDistance, 15)),
				}
			}),
			s.killMonsterByName(npc.Duriel, data.MonsterTypeNone, lightningSorceressMaxDistance, true, nil),
		}
	})
}

func (s LightningSorceress) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, lightningSorceressMaxDistance, false, skipOnImmunities)
}

func (s LightningSorceress) KillMephisto() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, 3, step.Distance(lightningSorceressMaxDistance, 15)),
				}
			}),
			s.killMonsterByName(npc.Mephisto, data.MonsterTypeNone, lightningSorceressMaxDistance, true, nil),
		}
	})
}

func (s LightningSorceress) KillNihlathak() action.Action {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, lightningSorceressMaxDistance, false, nil)
}

func (s LightningSorceress) KillDiablo() action.Action {
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
					step.SecondaryAttack(skill.StaticField, diablo.UnitID, 3, step.Distance(3, 8)),
				}
			}),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s LightningSorceress) KillIzual() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, 3, step.Distance(5, 8)),
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

func (s LightningSorceress) KillBaal() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d game.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(skill.StaticField, m.UnitID, 3, step.Distance(5, 8)),
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

func (s LightningSorceress) KillCouncil() action.Action {
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
	}, nil, step.Distance(8, lightningSorceressMaxDistance))
}

func (s LightningSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, maxDistance int, _ bool, skipOnImmunities []stat.Resist) action.Action {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities, step.Distance(lightningSorceressMinDistance, maxDistance))
}

func (s LightningSorceress) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}
