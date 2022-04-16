package character

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
	"strings"
)

const (
	hammerdinMaxAttacksLoop = 10
)

type Hammerdin struct {
	BaseCharacter
}

func (s Hammerdin) Buff() *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
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

func (s Hammerdin) KillCountess() *action.BasicAction {
	return s.killMonster(game.Countess)
}

func (s Hammerdin) KillAndariel() *action.BasicAction {
	return s.killMonster(game.Andariel)
}

func (s Hammerdin) KillSummoner() *action.BasicAction {
	return s.killMonster(game.Summoner)
}

func (s Hammerdin) KillPindle() *action.BasicAction {
	return s.killMonster(game.Pindleskin)
}

func (s Hammerdin) KillMephisto() *action.BasicAction {
	return s.killMonster(game.Mephisto)
}

func (s Hammerdin) KillNihlathak() *action.BasicAction {
	return s.killMonster(game.Nihlathak)
}

func (s Hammerdin) ClearAncientTunnels() *action.BasicAction {
	// Let's focus only on elite packs
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		var eliteMonsters []game.Monster
		for _, m := range data.Monsters {
			if m.Type == game.MonsterTypeMinion || m.Type == game.MonsterTypeUnique || m.Type == game.MonsterTypeChampion {
				eliteMonsters = append(eliteMonsters, m)
			}
		}

		sort.Slice(eliteMonsters, func(i, j int) bool {
			distanceI := pather.DistanceFromPoint(data, eliteMonsters[i].Position.X, eliteMonsters[i].Position.Y)
			distanceJ := pather.DistanceFromPoint(data, eliteMonsters[j].Position.X, eliteMonsters[j].Position.Y)

			return distanceI > distanceJ
		})

		for _, m := range eliteMonsters {
			for i := 0; i < hammerdinMaxAttacksLoop; i++ {
				steps = append(steps,
					step.PrimaryAttack(
						game.NPCID(m.Name),
						8,
						config.Config.Runtime.CastDuration,
						step.FollowEnemy(3),
						step.EnsureAura(config.Config.Bindings.Hammerdin.Concentration),
					),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s Hammerdin) KillCouncil() *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		// Exclude monsters that are not council members
		var councilMembers []game.Monster
		for _, m := range data.Monsters {
			if !strings.Contains(strings.ToLower(m.Name), "councilmember") {
				continue
			}
			councilMembers = append(councilMembers, m)
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := pather.DistanceFromPoint(data, councilMembers[i].Position.X, councilMembers[i].Position.Y)
			distanceJ := pather.DistanceFromPoint(data, councilMembers[j].Position.X, councilMembers[j].Position.Y)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			for i := 0; i < hammerdinMaxAttacksLoop; i++ {
				steps = append(steps,
					step.PrimaryAttack(
						game.NPCID(m.Name),
						8,
						config.Config.Runtime.CastDuration,
						step.FollowEnemy(3),
						step.EnsureAura(config.Config.Bindings.Hammerdin.Concentration),
					),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s Hammerdin) killMonster(npc game.NPCID) *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		helper.Sleep(100)
		for i := 0; i < hammerdinMaxAttacksLoop; i++ {
			steps = append(steps,
				step.PrimaryAttack(
					npc,
					8,
					config.Config.Runtime.CastDuration,
					step.FollowEnemy(3),
					step.EnsureAura(config.Config.Bindings.Hammerdin.Concentration),
				),
			)
		}

		return
	}, action.CanBeSkipped())
}
