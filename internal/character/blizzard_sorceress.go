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

const (
	sorceressMaxAttacksLoop = 10
	sorceressMinDistance    = 15
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) Buff() *action.BasicAction {
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

func (s BlizzardSorceress) KillCountess() *action.BasicAction {
	return s.killMonster(game.Countess, 20, true)
}

func (s BlizzardSorceress) KillAndariel() *action.BasicAction {
	return s.killMonster(game.Andariel, 20, false)
}

func (s BlizzardSorceress) KillSummoner() *action.BasicAction {
	return s.killMonster(game.Summoner, 10, false)
}

func (s BlizzardSorceress) KillPindle() *action.BasicAction {
	return s.killMonster(game.Pindleskin, 30, false)
}

func (s BlizzardSorceress) KillMephisto() *action.BasicAction {
	return s.killMonster(game.Mephisto, 20, true)
}

func (s BlizzardSorceress) KillNihlathak() *action.BasicAction {
	return s.killMonster(game.Nihlathak, 20, false)
}

func (s BlizzardSorceress) ClearAncientTunnels() *action.BasicAction {
	return nil
}

func (s BlizzardSorceress) KillCouncil() *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		// Exclude monsters that are not council members
		var councilMembers []game.Monster
		var coldImmunes []game.Monster
		for _, m := range data.Monsters {
			if !strings.Contains(strings.ToLower(m.Name), "councilmember") {
				continue
			}
			if m.IsImmune(game.ResistCold) {
				coldImmunes = append(coldImmunes, m)
			} else {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := pather.DistanceFromPoint(data, councilMembers[i].Position.X, councilMembers[i].Position.Y)
			distanceJ := pather.DistanceFromPoint(data, councilMembers[j].Position.X, councilMembers[j].Position.Y)

			return distanceI < distanceJ
		})

		councilMembers = append(councilMembers, coldImmunes...)

		for _, m := range councilMembers {
			for i := 0; i < sorceressMaxAttacksLoop; i++ {
				// Try to move closer after few attacks
				maxDistance := 30
				if i > 3 {
					maxDistance = 0
				}

				steps = append(steps,
					step.NewSecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, game.NPCID(m.Name), 1, time.Second, step.Distance(0, maxDistance)),
					step.PrimaryAttack(game.NPCID(m.Name), 4, config.Config.Runtime.CastDuration, step.Distance(0, maxDistance)),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s BlizzardSorceress) killMonster(npc game.NPCID, maxDistance int, useStaticField bool) *action.BasicAction {
	return action.BuildOnRuntime(func(data game.Data) (steps []step.Step) {
		if useStaticField {
			steps = append(steps, step.NewSecondaryAttack(config.Config.Bindings.Sorceress.StaticField, npc, 5, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, 15)))
		}

		for i := 0; i < sorceressMaxAttacksLoop; i++ {
			steps = append(steps,
				step.NewSecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, npc, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, maxDistance)),
				step.PrimaryAttack(npc, 4, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, maxDistance)),
			)
			if i == 1 {
				steps = append(steps,
					step.SyncStep(func(data game.Data) error {
						hid.MovePointer(hid.GameAreaSizeX/2, hid.GameAreaSizeY/2)
						hid.Click(hid.RightButton)
						return nil
					}),
				)
			}
		}

		return
	}, action.CanBeSkipped())
}
