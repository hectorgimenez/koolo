package character

import (
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type SorceressLeveling struct {
	BaseCharacter
}

func (s SorceressLeveling) GetSkillTree() skill.Tree {
	return skill.SorceressTree
}

func (s SorceressLeveling) ShouldResetSkills(d data.Data) bool {
	if d.PlayerUnit.Stats[stat.Level] >= 25 && d.PlayerUnit.Skills[skill.FireBall] > 10 {
		return true
	}

	return false
}

func (s SorceressLeveling) GetKeyBindings(d data.Data) map[skill.Skill]string {
	skillBindings := map[skill.Skill]string{
		skill.FrozenArmor:      config.Config.Bindings.Sorceress.FrozenArmor,
		skill.StaticField:      config.Config.Bindings.Sorceress.StaticField,
		skill.Teleport:         config.Config.Bindings.Teleport,
		skill.TomeOfTownPortal: config.Config.Bindings.TP,
	}

	if d.PlayerUnit.Skills[skill.Blizzard] > 0 {
		skillBindings[skill.Blizzard] = config.Config.Bindings.Sorceress.Blizzard
		skillBindings[skill.GlacialSpike] = ""
	} else if d.PlayerUnit.Skills[skill.FireBall] > 0 {
		skillBindings[skill.FireBall] = config.Config.Bindings.Sorceress.FireBall
	} else if d.PlayerUnit.Skills[skill.FireBolt] > 0 {
		skillBindings[skill.FireBolt] = config.Config.Bindings.Sorceress.FireBall
	}

	return skillBindings
}

func (s SorceressLeveling) StatPoints(d data.Data) map[stat.ID]int {
	if d.PlayerUnit.Stats[stat.Level] < 9 {
		return map[stat.ID]int{
			stat.Vitality: 9999,
		}
	}

	if d.PlayerUnit.Stats[stat.Level] < 15 {
		return map[stat.ID]int{
			stat.Energy:   45,
			stat.Strength: 25,
			stat.Vitality: 9999,
		}
	}

	return map[stat.ID]int{
		stat.Energy:   60,
		stat.Strength: 50,
		stat.Vitality: 9999,
	}
}

func (s SorceressLeveling) SkillPoints(d data.Data) []skill.Skill {
	if d.PlayerUnit.Stats[stat.Level] < 25 {
		return []skill.Skill{
			skill.FireBolt,
			skill.FrozenArmor,
			skill.FireBolt,
			skill.FireBolt,
			skill.Warmth,
			skill.StaticField,
			skill.FireBolt,
			skill.FireBolt,
			skill.FireBolt,
			skill.FireBolt,
			skill.Telekinesis,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.Teleport,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireBall,
			skill.FireMastery,
		}
	}

	return []skill.Skill{
		skill.StaticField,
		skill.StaticField,
		skill.StaticField,
		skill.StaticField,
		skill.Telekinesis,
		skill.Teleport,
		skill.FrozenArmor,
		skill.IceBolt,
		skill.IceBlast,
		skill.FrostNova,
		skill.GlacialSpike,
		skill.Blizzard,
		skill.Blizzard,
		skill.Warmth,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.IceBlast,
		skill.IceBlast,
		skill.IceBlast,
		skill.IceBlast,
		skill.IceBlast,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.ColdMastery,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.Blizzard,
		skill.ColdMastery,
		skill.ColdMastery,
		skill.ColdMastery,
		skill.ColdMastery,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
		skill.GlacialSpike,
	}
}

func (s SorceressLeveling) EnsureStatPoints() action.Action {
	return action.NewStepChain(func(d data.Data) []step.Step {
		_, found := d.PlayerUnit.Stats[stat.StatPoints]
		if !found {
			return []step.Step{}
		}

		return nil
	})
}

func (s SorceressLeveling) EnsureSkillPoints() action.Action {
	return action.NewStepChain(func(d data.Data) []step.Step {
		_, found := d.PlayerUnit.Stats[stat.SkillPoints]
		if !found {
			return []step.Step{}
		}

		return nil
	})
}

func (s SorceressLeveling) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillAndariel() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, m.UnitID, 1, step.Distance(25, 30)),
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(3, 5)),
				}
			}),
			s.killMonster(npc.Andariel, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLeveling) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s SorceressLeveling) KillDuriel() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Duriel, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLeveling) KillMephisto() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		// Let's try to moat trick if Teleport is available
		//if step.CanTeleport(d) {
		//	moatTrickPosition := data.Position{X: 17611, Y: 8093}
		//	return []action.Action{
		//		action.NewStepChain(func(d data.Data) []step.Step {
		//			mephisto, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
		//			return []step.Step{
		//				step.Wait(time.Second * 2),
		//				step.MoveTo(data.Position{X: 17580, Y: 8085}),
		//				step.Wait(time.Second * 3),
		//				step.MoveTo(moatTrickPosition),
		//				step.Wait(time.Second * 3),
		//				step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, mephisto.UnitID, 3),
		//			}
		//		}),
		//	}
		//}

		// If teleport is not available, just try to kill him with Static Field and Fire Ball
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				mephisto, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, mephisto.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Mephisto, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLeveling) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s SorceressLeveling) KillCouncil() action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
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

		if len(councilMembers) > 0 {
			return councilMembers[0].UnitID, true
		}

		return 0, false
	}, nil)
}

func (s SorceressLeveling) KillDiablo() action.Action {
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
			action.NewStepChain(func(d data.Data) []step.Step {
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, diablo.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
				}
			}),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s SorceressLeveling) KillIzual() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				monster, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, monster.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
				}
			}),
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLeveling) KillBaal() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.NewStepChain(func(d data.Data) []step.Step {
				baal, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, baal.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
				}
			}),
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
		}
	})
}

func (s SorceressLeveling) KillAncients() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
			actions = append(actions,
				action.NewStepChain(func(d data.Data) []step.Step {
					m, _ := d.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)
					return []step.Step{
						step.SecondaryAttack(config.Config.Bindings.Sorceress.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(8, 10)),
						step.MoveTo(data.Position{
							X: 10062,
							Y: 12639,
						}),
					}
				}),
				s.killMonster(m.Name, data.MonsterTypeSuperUnique),
			)
		}
		return actions
	})
}

func (s SorceressLeveling) KillMonsterSequence(monsterSelector func(d data.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
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

		if len(opts) == 0 {
			opts = append(opts, step.Distance(1, 30))
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		// During early game stages amount of mana is ridiculous...
		if d.PlayerUnit.MPPercent() < 15 && d.PlayerUnit.Stats[stat.Level] < 15 {
			steps = append(steps, step.PrimaryAttack(id, 1, step.Distance(1, 3)))
		} else {
			if _, found := d.PlayerUnit.Skills[skill.Blizzard]; found {
				if completedAttackLoops%2 == 0 {
					for _, m := range d.Monsters.Enemies() {
						if d := pather.DistanceFromMe(d, m.Position); d < 4 {
							s.logger.Debug("Monster detected close to the player, casting Blizzard over it")
							steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, m.UnitID, 1, step.Distance(25, 30)))
							break
						}
					}
				}

				steps = append(steps,
					step.SecondaryAttack(config.Config.Bindings.Sorceress.Blizzard, id, 1, step.Distance(25, 30)),
					step.PrimaryAttack(id, 3, step.Distance(25, 30)),
				)
			} else {
				steps = append(steps, step.SecondaryAttack(config.Config.Bindings.Sorceress.FireBall, id, 4, step.Distance(1, 25)))
			}
		}

		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s SorceressLeveling) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return s.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s SorceressLeveling) BuffSkills() map[skill.Skill]string {
	return map[skill.Skill]string{
		skill.FrozenArmor: config.Config.Bindings.Sorceress.FrozenArmor,
	}
}

func (s SorceressLeveling) staticFieldCasts() int {
	switch config.Config.Game.Difficulty {
	case difficulty.Normal:
		return 8
	}

	return 6
}
