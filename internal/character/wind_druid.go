package character

import (
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	druMaxAttacksLoop = 20
	druMinDistance    = 2
	druMaxDistance    = 8
)

var lastRavenAt = map[string]time.Time{}

type WindDruid struct {
	BaseCharacter
}

func (s WindDruid) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.Hurricane, skill.OakSage, skill.CycloneArmor, skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := d.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (du WindDruid) KillMonsterSequence(
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

		if !du.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		du.RecastBuffs(d)

		if completedAttackLoops >= druMaxAttacksLoop {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)
		// Add a random movement, maybe tornado is not hitting the target
		if previousUnitID == int(id) {
			steps = append(steps,
				step.SyncStep(func(d game.Data) error {
					monster, f := d.Monsters.FindByID(id)
					if f && monster.Stats[stat.Life] > 0 {
						du.container.PathFinder.RandomMovement(d)
					}
					return nil
				}),
			)
		}

		steps = append(steps,
			step.PrimaryAttack(
				id,
				3,
				true,
				step.Distance(druMinDistance, druMaxDistance),
			),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (du WindDruid) RecastBuffs(d game.Data) {
	skills := []skill.ID{skill.Hurricane, skill.OakSage, skill.CycloneArmor}
	states := []state.State{state.Hurricane, state.Oaksage, state.Cyclonearmor}

	for i, druSkill := range skills {
		if kb, found := d.KeyBindings.KeyBindingForSkill(druSkill); found {
			if !d.PlayerUnit.States.HasState(states[i]) {
				du.container.HID.PressKeyBinding(kb)
				helper.Sleep(180)
				du.container.HID.Click(game.RightButton, 640, 340)
				helper.Sleep(100)
			}
		}
	}
}

func (du WindDruid) BuffSkills(d game.Data) (buffs []skill.ID) {
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.CycloneArmor); found {
		buffs = append(buffs, skill.CycloneArmor)
	}
	if _, ravenFound := d.KeyBindings.KeyBindingForSkill(skill.Raven); ravenFound {
		buffs = append(buffs, skill.Raven, skill.Raven, skill.Raven, skill.Raven, skill.Raven)
	}
	if _, hurricaneFound := d.KeyBindings.KeyBindingForSkill(skill.Hurricane); hurricaneFound {
		buffs = append(buffs, skill.Hurricane)
	}
	return buffs
}

func (du WindDruid) PreCTABuffSkills(d game.Data) (skills []skill.ID) {
	needsBear := true
	wolves := 5
	direWolves := 3
	needsOak := true
	needsCreeper := true

	for _, monster := range d.Monsters {
		if monster.IsPet() {
			if monster.Name == npc.DruBear {
				needsBear = false
			}
			if monster.Name == npc.DruFenris {
				direWolves--
			}
			if monster.Name == npc.DruSpiritWolf {
				wolves--
			}
			if monster.Name == npc.OakSage {
				needsOak = false
			}
			if monster.Name == npc.DruCycleOfLife {
				needsCreeper = false
			}
			if monster.Name == npc.VineCreature {
				needsCreeper = false
			}
			if monster.Name == npc.DruPlaguePoppy {
				needsCreeper = false
			}
		}
	}

	if d.PlayerUnit.States.HasState(state.Oaksage) {
		needsOak = false
	}

	_, foundDireWolf := d.KeyBindings.KeyBindingForSkill(skill.SummonDireWolf)
	_, foundWolf := d.KeyBindings.KeyBindingForSkill(skill.SummonSpiritWolf)
	_, foundBear := d.KeyBindings.KeyBindingForSkill(skill.SummonGrizzly)
	_, foundOak := d.KeyBindings.KeyBindingForSkill(skill.OakSage)
	_, foundSolarCreeper := d.KeyBindings.KeyBindingForSkill(skill.SolarCreeper)
	_, foundCarrionCreeper := d.KeyBindings.KeyBindingForSkill(skill.CarrionVine)
	_, foundPoisonCreeper := d.KeyBindings.KeyBindingForSkill(skill.PoisonCreeper)

	if foundWolf {
		for i := 0; i < wolves; i++ {
			skills = append(skills, skill.SummonSpiritWolf)
		}
	}
	if foundDireWolf {
		for i := 0; i < direWolves; i++ {
			skills = append(skills, skill.SummonDireWolf)
		}
	}
	if foundBear && needsBear {
		skills = append(skills, skill.SummonGrizzly)
	}
	if foundOak && needsOak {
		skills = append(skills, skill.OakSage)
	}
	if foundSolarCreeper && needsCreeper {
		skills = append(skills, skill.SolarCreeper)
	}
	if foundCarrionCreeper && needsCreeper {
		skills = append(skills, skill.CarrionVine)
	}
	if foundPoisonCreeper && needsCreeper {
		skills = append(skills, skill.PoisonCreeper)
	}

	return skills
}

func (du WindDruid) KillCountess() action.Action {
	return du.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (du WindDruid) KillAndariel() action.Action {
	return du.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (du WindDruid) KillSummoner() action.Action {
	return du.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (du WindDruid) KillDuriel() action.Action {
	return du.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (du WindDruid) KillPindle(_ []stat.Resist) action.Action {
	return du.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (du WindDruid) KillMephisto() action.Action {
	return du.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (du WindDruid) KillNihlathak() action.Action {
	return du.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (du WindDruid) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			du.logger.Error("Diablo was not found, timeout reached")
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
		du.logger.Info("Diablo detected, attacking")

		return []action.Action{
			du.killMonster(npc.Diablo, data.MonsterTypeNone),
			du.killMonster(npc.Diablo, data.MonsterTypeNone),
			du.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (du WindDruid) KillIzual() action.Action {
	return du.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (du WindDruid) KillCouncil() action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := pather.DistanceFromMe(d, councilMembers[i].Position)
			distanceJ := pather.DistanceFromMe(d, councilMembers[j].Position)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			for range druMaxAttacksLoop {
				steps = append(steps,
					step.PrimaryAttack(
						m.UnitID,
						3,
						true,
						step.Distance(druMinDistance, druMaxDistance),
					),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (du WindDruid) KillBaal() action.Action {
	return du.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}

func (du WindDruid) killMonster(npcId npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npcId, t)
		if !found {
			return nil
		}

		helper.Sleep(100)
		du.RecastBuffs(d)
		for range druMaxAttacksLoop {
			steps = append(steps,
				step.PrimaryAttack(
					m.UnitID,
					3,
					true,
					step.Distance(druMinDistance, druMaxDistance),
				),
			)
		}
		return
	}, action.CanBeSkipped())
}
