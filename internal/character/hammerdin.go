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
	"time"
)

type Hammerdin struct {
	BaseCharacter
}

func (s Hammerdin) Buff() *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		steps = append(steps, s.buffCTA()...)
		steps = append(steps, step.SyncStep(func(data game.Data) error {
			if config.Config.Bindings.Sorceress.FrozenArmor != "" {
				hid.PressKey(config.Config.Bindings.Sorceress.FrozenArmor)
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
	return s.killMonster(game.TheSummoner)
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

		// Order council members by distance and immunities
		sort.Slice(councilMembers, func(i, j int) bool {
			if councilMembers[j].IsImmune(game.ResistCold) {
				return false
			}

			distanceI := pather.DistanceFromPoint(data, councilMembers[i].Position.X, councilMembers[i].Position.Y)
			distanceJ := pather.DistanceFromPoint(data, councilMembers[j].Position.X, councilMembers[j].Position.Y)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			for i := 0; i < sorceressMaxAttacksLoop; i++ {
				// Try to move closer after few attacks
				maxDistance := 20
				if i > 3 {
					maxDistance = 0
				}

				steps = append(steps,
					step.NewSecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, game.NPCID(m.Name), 1, time.Second, step.FollowEnemy(maxDistance)),
					step.PrimaryAttack(game.NPCID(m.Name), 4, config.Config.Runtime.CastDuration, step.FollowEnemy(maxDistance)),
				)
			}
		}
		return
	})

}

func (s Hammerdin) killMonster(npc game.NPCID) *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		for i := 0; i < sorceressMaxAttacksLoop; i++ {
			steps = append(steps,
				step.NewSecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, npc, 1, time.Second),
				step.PrimaryAttack(npc, 4, config.Config.Runtime.CastDuration),
			)
		}

		return
	})
}
