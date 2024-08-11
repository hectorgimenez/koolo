package character

import (
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	hammerdinMaxAttacksLoop = 20 // Adjust from 5-20 depending on DMG and rotation, lower attack loops would cause higher attack rotation whereas bigger would perform multiple(longer) attacks on one spot.
)

type Hammerdin struct {
	BaseCharacter
}

func (s Hammerdin) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.Concentration, skill.HolyShield, skill.TomeOfTownPortal}
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

func (s Hammerdin) KillMonsterSequence(
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

		if completedAttackLoops >= hammerdinMaxAttacksLoop {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)
		// Add a random movement, maybe hammer is not hitting the target
		if previousUnitID == int(id) {
			steps = append(steps,
				step.SyncStep(func(d game.Data) error {
					monster, f := d.Monsters.FindByID(id)
					if f && monster.Stats[stat.Life] > 0 {
						s.container.PathFinder.RandomMovement(d)
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
				step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
				step.EnsureAura(skill.Concentration),
			),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s Hammerdin) BuffSkills(d game.Data) []skill.ID {
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return []skill.ID{}
}

func (s Hammerdin) PreCTABuffSkills(_ game.Data) []skill.ID {
	return []skill.ID{}
}

func (s Hammerdin) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillAndariel() action.Action {
	return s.killBoss(npc.Andariel, data.MonsterTypeNone)
}

func (s Hammerdin) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s Hammerdin) KillDuriel() action.Action {
	return s.killBoss(npc.Duriel, data.MonsterTypeNone)
}

func (s Hammerdin) KillPindle(_ []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillMephisto() action.Action {
	return s.killBoss(npc.Mephisto, data.MonsterTypeNone)
}

func (s Hammerdin) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillDiablo() action.Action {
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
			s.killBoss(npc.Diablo, data.MonsterTypeNone),
			s.killBoss(npc.Diablo, data.MonsterTypeNone),
			s.killBoss(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s Hammerdin) KillIzual() action.Action {
	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s Hammerdin) KillCouncil() action.Action {
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
			for range hammerdinMaxAttacksLoop {
				steps = append(steps,
					step.PrimaryAttack(
						m.UnitID,
						8,
						true,
						step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
						step.EnsureAura(skill.Concentration),
					),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s Hammerdin) KillBaal() action.Action {
	return s.killBoss(npc.BaalCrab, data.MonsterTypeNone)
}

func (s Hammerdin) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}

		helper.Sleep(100)
		for range hammerdinMaxAttacksLoop {
			steps = append(steps,
				step.PrimaryAttack(
					m.UnitID,
					8,
					true,
					step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
					step.EnsureAura(skill.Concentration),
				),
				step.SyncStep(func(d game.Data) error {
					m, found = d.Monsters.FindOne(m.Name, t)
					if found && m.Stats[stat.Life] > 0 {
						s.container.PathFinder.RandomMovement(d)
					}
					return nil
				}),
			)
		}

		return
	}, action.CanBeSkipped())
}

func (s Hammerdin) killBoss(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}

		helper.Sleep(100)
		for range hammerdinMaxAttacksLoop {
			steps = append(steps,
				step.PrimaryAttack(
					m.UnitID,
					8,
					true,
					step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
					step.EnsureAura(skill.Concentration),
				),
			)
		}

		return
	}, action.CanBeSkipped())
}
