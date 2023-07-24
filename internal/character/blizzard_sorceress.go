package character

import (
	"fmt"
	"math"
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
	sorceressMaxAttacksLoop = 10
	sorceressMinDistance    = 25
	sorceressMaxDistance    = 30
)

type BlizzardSorceress struct {
	BaseCharacter
}

func (s BlizzardSorceress) KillMonsterSequence(
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
			opts = append(opts, step.Distance(sorceressMinDistance, sorceressMaxDistance))
		}
		//if useStaticField {
		//	steps = append(steps,
		//		step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, time.Millisecond*100, step.Distance(sorceressMinDistance, maxDistance)),
		//		step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, id, 5, config.Config.Runtime.CastDuration, step.Distance(sorceressMinDistance, 15)),
		//	)
		//}
		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}, false
		}

		steps := make([]step.Step, 0)
		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		if completedAttackLoops%2 == 0 {
			for _, m := range d.Monsters.Enemies() {
				if d := pather.DistanceFromMe(d, m.Position); d < 4 {
					s.logger.Debug("Monster detected close to the player, casting Blizzard over it")
					steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, m.UnitID, 1, opts...))
					break
				}
			}
		}

		// In case monster is stuck behind a wall or character is not able to reachh it we will short the distance
		if completedAttackLoops > 5 {
			s.logger.Debug("Looks like monster is not reachable, moving closer")
			opts = append(opts, step.Distance(2, 8))
		}

		steps = append(steps,
			step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, opts...),
			step.PrimaryAttack(id, 4, opts...),
		)
		completedAttackLoops++
		previousUnitID = int(id)

		return steps, true
	})
}

func (s BlizzardSorceress) CastProtectiveSpells() action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
		near_by_count :=0
		near_by_distance := 8
		var nearest_moster *data.Monster
		nearest_distance := math.MaxInt

		monsters := d.Monsters.Enemies()
		for i, m := range monsters {
			distance := pather.DistanceFromMe(d, m.Position)
			if distance < near_by_distance {
				near_by_count++
			}
			if distance < nearest_distance {
				nearest_distance = distance
				nearest_moster = &monsters[i]
			}
			// fmt.Println(fmt.Sprintf( "enemy [%d] id [%d] at x:[%d] y:[%d], type:[%s], distance:[%d]", m.Name, m.UnitID, m.Position.X, m.Position.Y, m.Type, distance))
		}

		if near_by_count > 3 {
			s.logger.Debug(fmt.Sprintf("Nearby enemy more than %d, cast frost nova", near_by_count))
			steps = append(steps, step.SyncStep(func(d data.Data) error {
				if config.Config.Bindings.Sorceress.FrostNova != "" {
					hid.PressKey(config.Config.Bindings.Sorceress.FrostNova)
					helper.Sleep(100)
					hid.Click(hid.RightButton)
				}

				return nil
			}))
		}

		// fmt.Println(fmt.Sprintf( "nearest monster [%d] id [%d] distance:[%d]", nearest_moster.Name, nearest_moster.UnitID, nearest_distance))
		
		steps = append(steps,
			step.CastAt(config.Config.Bindings.Sorceress.Blizzard, nearest_moster.Position, 1),
		)

		return
	})
}

func (s BlizzardSorceress) Buff() action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
		steps = append(steps, s.buffCTA()...)
		steps = append(steps, step.SyncStep(func(d data.Data) error {
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
	return action.NewChain(func(d data.Data) (actions []action.Action){
		actions = append(actions,
			s.CastProtectiveSpells(),
			s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil),
		)
		return
	}) 
}

func (s BlizzardSorceress) KillAndariel() action.Action {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillSummoner() action.Action {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeNone, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, skipOnImmunities)
}

func (s BlizzardSorceress) KillMephisto() action.Action {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeNone, sorceressMaxDistance, true, nil)
}

func (s BlizzardSorceress) KillNihlathak() action.Action {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, nil)
}

func (s BlizzardSorceress) KillCouncil() action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var coldImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
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
			return m.UnitID, true
		}

		return 0, false
	}, nil, step.Distance(8, sorceressMaxDistance))
}

func (s BlizzardSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, maxDistance int, useStaticField bool, skipOnImmunities []stat.Resist) action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities, step.Distance(sorceressMinDistance, maxDistance))
}
