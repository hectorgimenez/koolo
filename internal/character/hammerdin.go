package character

import (
	"sort"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
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
) *action.DynamicAction {
	completedAttackLoops := 0
	previousUnitID := 0

	return action.BuildDynamic(func(d data.Data) ([]step.Step, bool) {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}, false
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}, false
		}

		if len(opts) == 0 {
			opts = append(opts, step.Distance(1, 5))
		}

		if completedAttackLoops >= hammerdinMaxAttacksLoop {
			return []step.Step{}, false
		}

		steps := make([]step.Step, 0)
		steps = append(steps,
			step.PrimaryAttack(
				id,
				8,
				step.Distance(2, 8),
				step.EnsureAura(config.Config.Bindings.Hammerdin.Concentration),
			),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps, true
	})
}

func (s Hammerdin) Buff() action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
		steps = append(steps, s.buffCTA()...)
		steps = append(steps, step.SyncStep(func(d data.Data) error {
			if config.Config.Bindings.Hammerdin.HolyShield != "" {
				hid.PressKey(config.Config.Bindings.Hammerdin.HolyShield)
				helper.Sleep(100)
				hid.Click(hid.RightButton)
			}

			return nil
		}))

		return
	})
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
	panic("implement me")
}

func (s Hammerdin) KillIzual() action.Action {
	panic("implement me")
}

func (s Hammerdin) KillCouncil() action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
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
						step.EnsureAura(config.Config.Bindings.Hammerdin.Concentration),
					),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s Hammerdin) KillBaal() action.Action {
	//TODO implement me
	panic("implement me")
}

func (s Hammerdin) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
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
					step.EnsureAura(config.Config.Bindings.Hammerdin.Concentration),
				),
			)
		}

		return
	}, action.CanBeSkipped())
}
