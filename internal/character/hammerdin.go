package character

import (
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	hammerdinMaxAttacksLoop = 20
)

type Hammerdin struct {
	BaseCharacter
}

func (s Hammerdin) KillMonsterSequence(
	monsterSelector func(d data.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0

	return action.NewStepChain(func(d data.Data) []step.Step {
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
				step.SyncStep(func(d data.Data) error {
					monster, f := d.Monsters.FindByID(id)
					if f && monster.Stats[stat.Life] > 0 {
						pather.RandomMovement()
					}
					return nil
				}),
			)
		}
		steps = append(steps,
			step.PrimaryAttack(
				id,
				8,
				step.Distance(2, 8),
				step.EnsureAura(config.Config.Bindings.Paladin.Concentration),
			),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s Hammerdin) BuffSkills() map[skill.Skill]string {
	return map[skill.Skill]string{
		skill.HolyShield: config.Config.Bindings.Paladin.HolyShield,
	}
}

func (s Hammerdin) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s Hammerdin) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s Hammerdin) KillDuriel() action.Action {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s Hammerdin) KillPindle(_ []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s Hammerdin) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d data.Data) []action.Action {
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
			return []action.Action{action.NewStepChain(func(d data.Data) []step.Step {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			})}
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return []action.Action{
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s Hammerdin) KillIzual() action.Action {
	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s Hammerdin) KillCouncil() action.Action {
	return action.NewStepChain(func(d data.Data) (steps []step.Step) {
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
			for i := 0; i < hammerdinMaxAttacksLoop; i++ {
				steps = append(steps,
					step.PrimaryAttack(
						m.UnitID,
						8,
						step.Distance(2, 8),
						step.EnsureAura(config.Config.Bindings.Paladin.Concentration),
					),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s Hammerdin) KillBaal() action.Action {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}

func (s Hammerdin) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d data.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}

		helper.Sleep(100)
		for i := 0; i < hammerdinMaxAttacksLoop; i++ {
			steps = append(steps,
				step.PrimaryAttack(
					m.UnitID,
					8,
					step.Distance(2, 8),
					step.EnsureAura(config.Config.Bindings.Paladin.Concentration),
				),
			)
		}

		return
	}, action.CanBeSkipped())
}
