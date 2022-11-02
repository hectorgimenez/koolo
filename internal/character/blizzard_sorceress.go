package character

import (
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/game/npc"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
	"time"
)

const (
	sorceressMaxAttacksLoop = 10
	sorceressMinDistance    = 15
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) Buff() action.Action {
	return action.BuildStatic(func(data game.Data) (steps []step.Step) {
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

func (s BlizzardSorceress) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, game.MonsterTypeSuperUnique, 20, true)
}

func (s BlizzardSorceress) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, game.MonsterTypeNone, 20, false)
}

func (s BlizzardSorceress) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, game.MonsterTypeNone, 10, false)
}

func (s BlizzardSorceress) KillPindle() action.Action {
	return s.killMonster(npc.DefiledWarrior, game.MonsterTypeSuperUnique, 30, false)
}

func (s BlizzardSorceress) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, game.MonsterTypeNone, 20, true)
}

func (s BlizzardSorceress) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, game.MonsterTypeSuperUnique, 20, false)
}

func (s BlizzardSorceress) ClearAncientTunnels() action.Action {
	return nil
}

func (s BlizzardSorceress) KillCouncil() action.Action {
	toggleSeconday := true
	return action.BuildDynamic(func(data game.Data) (step.Step, bool) {
		// Exclude monsters that are not council members
		var councilMembers []game.Monster
		var coldImmunes []game.Monster
		for _, m := range data.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(game.ResistCold) {
					coldImmunes = append(coldImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
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
			if toggleSeconday {
				toggleSeconday = false
				return step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, m.Name, 1, time.Second, step.Distance(0, 30)), true
			}

			toggleSeconday = true
			return step.PrimaryAttack(m.Name, 4, config.Config.Runtime.CastDuration, step.Distance(0, 30)), true
		}

		return nil, false
	}, action.CanBeSkipped())
}

func (s BlizzardSorceress) killMonster(npc npc.ID, t game.MonsterType, maxDistance int, useStaticField bool) action.Action {
	return action.BuildStatic(func(data game.Data) (steps []step.Step) {
		if useStaticField {
			steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, npc, 5, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, 15), step.MonsterType(t)))
		}

		for i := 0; i < sorceressMaxAttacksLoop; i++ {
			steps = append(steps,
				step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, npc, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, maxDistance), step.MonsterType(t)),
				step.PrimaryAttack(npc, 4, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, maxDistance), step.MonsterType(t)),
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
