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
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/hid"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type SorceressLeveling struct {
	BaseCharacter
}

func (s SorceressLeveling) GetKeyBindings() map[skill.Skill]string {
	return map[skill.Skill]string{
		skill.ChargedBolt:      config.Config.Bindings.SorceressLeveling.Nova,
		skill.Nova:             config.Config.Bindings.SorceressLeveling.Nova,
		skill.Blizzard:         config.Config.Bindings.SorceressLeveling.Nova,
		skill.FrozenArmor:      config.Config.Bindings.SorceressLeveling.FrozenArmor,
		skill.FrostNova:        config.Config.Bindings.SorceressLeveling.FrostNova,
		skill.StaticField:      config.Config.Bindings.SorceressLeveling.StaticField,
		skill.Teleport:         config.Config.Bindings.Teleport,
		skill.TomeOfTownPortal: config.Config.Bindings.TP,
	}
}

func (s SorceressLeveling) StatPoints(d data.Data) map[stat.ID]int {
	if d.PlayerUnit.Stats[stat.Level] < 15 {
		return map[stat.ID]int{
			stat.Dexterity: 0,
			stat.Energy:    45,
			stat.Strength:  20,
			stat.Vitality:  9999,
		}
	}

	return map[stat.ID]int{
		stat.Dexterity: 0,
		stat.Energy:    60,
		stat.Strength:  45,
		stat.Vitality:  9999,
	}
}

func (s SorceressLeveling) SkillPoints() []skill.Skill {
	return []skill.Skill{
		skill.ChargedBolt,
		skill.ChargedBolt,
		skill.ChargedBolt,
		skill.ChargedBolt,
		skill.FrozenArmor,
		skill.FrostNova,
		skill.StaticField,
		skill.StaticField,
		skill.StaticField,
		skill.StaticField,
		skill.Telekinesis,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Teleport,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
		skill.Nova,
	}
}

func (s SorceressLeveling) EnsureStatPoints() action.Action {
	return action.BuildStatic(func(d data.Data) []step.Step {
		_, found := d.PlayerUnit.Stats[stat.StatPoints]
		if !found {
			return []step.Step{}
		}

		return nil
	})
}

func (s SorceressLeveling) EnsureSkillPoints() action.Action {
	return action.BuildStatic(func(d data.Data) []step.Step {
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
			action.BuildStatic(func(d data.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Andariel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, m.UnitID, 1, step.Distance(5, 10)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(3, 5)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, m.UnitID, 1, step.Distance(5, 10)),
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
			action.BuildStatic(func(d data.Data) []step.Step {
				m, _ := d.Monsters.FindOne(npc.Duriel, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, m.UnitID, 1, step.Distance(5, 10)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.StaticField, m.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, m.UnitID, 1, step.Distance(5, 10)),
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
		//		action.BuildStatic(func(d data.Data) []step.Step {
		//			mephisto, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
		//			return []step.Step{
		//				step.Wait(time.Second * 2),
		//				step.MoveTo(data.Position{X: 17580, Y: 8085}),
		//				step.Wait(time.Second * 3),
		//				step.MoveTo(moatTrickPosition),
		//				step.Wait(time.Second * 3),
		//				step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.Blizzard, mephisto.UnitID, 3),
		//			}
		//		}),
		//	}
		//}

		// If teleport is not available, just try to kill him with Static Field and Fire Ball
		return []action.Action{
			action.BuildStatic(func(d data.Data) []step.Step {
				mephisto, _ := d.Monsters.FindOne(npc.Mephisto, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, mephisto.UnitID, 1, step.Distance(5, 10)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.StaticField, mephisto.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, mephisto.UnitID, 1, step.Distance(5, 10)),
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
	startedAt := time.Time{}

	return action.NewFactory(func(d data.Data) action.Action {
		if startedAt.IsZero() {
			s.logger.Info("Waiting for Diablo to spawn")
			startedAt = time.Now()
		}

		if time.Since(startedAt) < timeout {
			return action.NewChain(func(d data.Data) []action.Action {
				diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
				if !found {
					return nil
				}

				s.logger.Info("Diablo detected, attacking")
				return []action.Action{
					action.BuildStatic(func(d data.Data) []step.Step {
						return []step.Step{
							step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, diablo.UnitID, 1, step.Distance(5, 10)),
							step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.StaticField, diablo.UnitID, s.staticFieldCasts(), step.Distance(1, 5)),
							step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, diablo.UnitID, 1, step.Distance(5, 10)),
						}
					}),
					s.killMonster(npc.Diablo, data.MonsterTypeNone),
				}
			})
		}

		s.logger.Info("Timeout waiting for Diablo to spawn")
		return nil
	})
}

func (s SorceressLeveling) KillIzual() action.Action {
	return action.NewChain(func(d data.Data) []action.Action {
		return []action.Action{
			action.BuildStatic(func(d data.Data) []step.Step {
				monster, _ := d.Monsters.FindOne(npc.Izual, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, monster.UnitID, 1, step.Distance(5, 10)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.StaticField, monster.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, monster.UnitID, 1, step.Distance(5, 10)),
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
			action.BuildStatic(func(d data.Data) []step.Step {
				baal, _ := d.Monsters.FindOne(npc.BaalCrab, data.MonsterTypeNone)
				return []step.Step{
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, baal.UnitID, 1, step.Distance(5, 10)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.StaticField, baal.UnitID, s.staticFieldCasts(), step.Distance(1, 4)),
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.FrostNova, baal.UnitID, 1, step.Distance(5, 10)),
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

func (s SorceressLeveling) KillMonsterSequence(monsterSelector func(d data.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) *action.DynamicAction {
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
			opts = append(opts, step.Distance(1, 30))
		}

		if completedAttackLoops >= sorceressMaxAttacksLoop {
			return []step.Step{}, false
		}

		steps := make([]step.Step, 0)

		// During early game stages amount of mana is ridiculous...
		if d.PlayerUnit.MPPercent() < 15 && d.PlayerUnit.Stats[stat.Level] < 15 {
			steps = append(steps, step.PrimaryAttack(id, 1, step.Distance(1, 3)))
		} else {
			m, _ := d.Monsters.FindByID(id)
			if _, found := d.PlayerUnit.Skills[skill.Blizzard]; found {
				steps = append(steps,
					step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.Blizzard, id, 1, step.Distance(1, 25)),
					step.PrimaryAttack(id, 3, step.Distance(1, 25)),
				)
			} else {
				maxDistance := 15
				if _, found := d.PlayerUnit.Skills[skill.Nova]; !found {
					// This means we are still using charged bolt
					maxDistance = 5
				}

				if m.IsElite() {
					steps = append(steps,
						step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.Nova, id, 3, step.Distance(1, maxDistance)),
					)
				} else {
					steps = append(steps,
						step.SecondaryAttack(config.Config.Bindings.SorceressLeveling.Nova, id, 1, step.Distance(1, maxDistance)),
					)
				}
			}
		}

		completedAttackLoops++
		previousUnitID = int(id)

		return steps, true
	})
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

func (s SorceressLeveling) Buff() action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
		return []step.Step{
			step.SyncStep(func(d data.Data) error {
				if _, found := d.PlayerUnit.Skills[skill.FrozenArmor]; !found {
					return nil
				}

				if config.Config.Bindings.SorceressLeveling.FrozenArmor != "" {
					hid.PressKey(config.Config.Bindings.SorceressLeveling.FrozenArmor)
					helper.Sleep(100)
					hid.Click(hid.RightButton)
				}

				return nil
			}),
		}
	})
}

func (s SorceressLeveling) staticFieldCasts() int {
	switch config.Config.Game.Difficulty {
	case difficulty.Normal:
		return 12
	}

	return 6
}
