package character

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/game/stat"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
)

const (
	hammerdinMaxAttacksLoop = 10
)

type Hammerdin struct {
	BaseCharacter
}

func (s Hammerdin) KillMonsterSequence(
	monsterSelector func(data game.Data) (game.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) *action.DynamicAction {
	//TODO implement me
	panic("implement me")
}

func (s Hammerdin) Buff() action.Action {
	return action.BuildStatic(func(data game.Data) (steps []step.Step) {
		steps = append(steps, s.buffCTA()...)
		steps = append(steps, step.SyncStep(func(data game.Data) error {
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
	return s.killMonster(npc.DarkStalker, game.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, game.MonsterTypeNone)
}

func (s Hammerdin) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, game.MonsterTypeNone)
}

func (s Hammerdin) KillPindle(_ []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, game.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, game.MonsterTypeNone)
}

func (s Hammerdin) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, game.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillCouncil() action.Action {
	return action.BuildStatic(func(data game.Data) (steps []step.Step) {
		// Exclude monsters that are not council members
		var councilMembers []game.Monster
		for _, m := range data.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := pather.DistanceFromMe(data, councilMembers[i].Position.X, councilMembers[i].Position.Y)
			distanceJ := pather.DistanceFromMe(data, councilMembers[j].Position.X, councilMembers[j].Position.Y)

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

func (s Hammerdin) killMonster(npc npc.ID, t game.MonsterType) action.Action {
	return action.BuildStatic(func(data game.Data) (steps []step.Step) {
		m, found := data.Monsters.FindOne(npc, t)
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
