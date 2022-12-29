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
	"time"
)

const (
	sorceressMaxAttacksLoop = 10
	sorceressMinDistance    = 15
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) KillMonsterSequence(data game.Data, id game.UnitID) (steps []step.Step) {
	if !s.preBattleChecks(data, id, []stat.Resist{}) {
		return
	}

	for i := 0; i < sorceressMaxAttacksLoop; i++ {
		steps = append(steps,
			step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, 10)),
			step.PrimaryAttack(id, 4, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, 10)),
		)
		if i == 1 {
			// Cast a Blizzard over character, to clear possible trash mobs
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
	return s.killMonsterByName(npc.DarkStalker, game.MonsterTypeSuperUnique, 20, true, nil)
}

func (s BlizzardSorceress) KillAndariel() action.Action {
	return s.killMonsterByName(npc.Andariel, game.MonsterTypeNone, 20, false, nil)
}

func (s BlizzardSorceress) KillSummoner() action.Action {
	return s.killMonsterByName(npc.Summoner, game.MonsterTypeNone, 10, false, nil)
}

func (s BlizzardSorceress) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonsterByName(npc.DefiledWarrior, game.MonsterTypeSuperUnique, 30, false, skipOnImmunities)
}

func (s BlizzardSorceress) KillMephisto() action.Action {
	return s.killMonsterByName(npc.Mephisto, game.MonsterTypeNone, 20, true, nil)
}

func (s BlizzardSorceress) KillNihlathak() action.Action {
	return s.killMonsterByName(npc.Nihlathak, game.MonsterTypeSuperUnique, 20, false, nil)
}

func (s BlizzardSorceress) KillCouncil() action.Action {
	return action.BuildDynamic(func(data game.Data) ([]step.Step, bool) {
		// Exclude monsters that are not council members
		var councilMembers []game.Monster
		var coldImmunes []game.Monster
		for _, m := range data.Monsters.Enemies() {
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
			return s.KillMonsterSequence(data, m.UnitID), true
		}

		return nil, false
	}, action.CanBeSkipped())
}

func (s BlizzardSorceress) killMonsterByName(id npc.ID, monsterType game.MonsterType, maxDistance int, useStaticField bool, skipOnImmunities []stat.Resist) action.Action {
	return action.BuildStatic(func(data game.Data) []step.Step {
		m, found := data.Monsters.FindOne(id, monsterType)
		if found {
			return s.killMonsterSteps(m.UnitID, maxDistance, useStaticField, skipOnImmunities)(data)
		}

		return nil
	})
}

func (s BlizzardSorceress) killMonsterSteps(id game.UnitID, maxDistance int, useStaticField bool, skipOnImmunities []stat.Resist) func(data game.Data) (steps []step.Step) {
	return func(data game.Data) (steps []step.Step) {
		if !s.preBattleChecks(data, id, skipOnImmunities) {
			return
		}

		if useStaticField {
			steps = append(steps,
				step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, maxDistance)),
				step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, id, 5, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, 15)),
			)
		}

		for i := 0; i < sorceressMaxAttacksLoop; i++ {
			steps = append(steps,
				step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, maxDistance)),
				step.PrimaryAttack(id, 4, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, maxDistance)),
			)
			if i == 1 {
				// Cast a Blizzard over character, to clear possible trash mobs
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
	}
}
