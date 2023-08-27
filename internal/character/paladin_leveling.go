package character

import (
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
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

type PaladinLeveling struct {
	BaseCharacter
}

func (p PaladinLeveling) GetSkillTree() skill.Tree {
	return skill.PaladinTree
}

func (p PaladinLeveling) Buff() action.Action {
	return action.BuildStatic(func(d data.Data) (steps []step.Step) {
		return []step.Step{
			step.SyncStep(func(d data.Data) error {
				if _, found := d.PlayerUnit.Skills[skill.HolyShield]; !found {
					return nil
				}

				if config.Config.Bindings.Paladin.HolyShield != "" {
					hid.PressKey(config.Config.Bindings.Paladin.HolyShield)
					helper.Sleep(100)
					hid.Click(hid.RightButton)
				}

				return nil
			}),
		}
	})
}

func (p PaladinLeveling) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return p.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (p PaladinLeveling) KillCountess() action.Action {
	return p.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (p PaladinLeveling) KillAndariel() action.Action {
	return p.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillSummoner() action.Action {
	return p.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillDuriel() action.Action {
	return p.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillMephisto() action.Action {
	return p.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillPindle(_ []stat.Resist) action.Action {
	return p.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (p PaladinLeveling) KillNihlathak() action.Action {
	return p.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (p PaladinLeveling) KillCouncil() action.Action {
	return p.KillMonsterSequence(func(d data.Data) (data.UnitID, bool) {
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

func (p PaladinLeveling) KillDiablo() action.Action {
	timeout := time.Second * 20
	startedAt := time.Time{}

	return action.NewFactory(func(d data.Data) action.Action {
		if startedAt.IsZero() {
			p.logger.Info("Waiting for Diablo to spawn")
			startedAt = time.Now()
		}

		if time.Since(startedAt) < timeout {
			return action.NewChain(func(d data.Data) []action.Action {
				_, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
				if !found {
					return nil
				}

				p.logger.Info("Diablo detected, attacking")
				return []action.Action{
					p.killMonster(npc.Diablo, data.MonsterTypeNone),
				}
			})
		}

		p.logger.Info("Timeout waiting for Diablo to spawn")
		return nil
	})
}

func (p PaladinLeveling) KillIzual() action.Action {
	return p.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillBaal() action.Action {
	return p.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillMonsterSequence(monsterSelector func(d data.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) *action.DynamicAction {
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

		if !p.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}, false
		}

		if completedAttackLoops >= 70 {
			return []step.Step{}, false
		}

		steps := make([]step.Step, 0)

		numOfAttacks := 3

		if d.PlayerUnit.Skills[skill.BlessedHammer] > 0 {
			steps = append(steps,
				step.PrimaryAttack(id, numOfAttacks, step.Distance(3, 5), step.EnsureAura(config.Config.Bindings.Paladin.Concentration)),
			)
		} else {
			if d.PlayerUnit.Skills[skill.Zeal] > 0 {
				numOfAttacks = 1
			}

			steps = append(steps,
				step.PrimaryAttack(id, numOfAttacks, step.Distance(1, 3), step.EnsureAura(config.Config.Bindings.Paladin.Concentration)),
			)
		}

		completedAttackLoops++
		previousUnitID = int(id)
		return steps, true
	})
}

func (p PaladinLeveling) StatPoints(d data.Data) map[stat.ID]int {
	if d.PlayerUnit.Stats[stat.Level] >= 21 && d.PlayerUnit.Stats[stat.Level] < 30 {
		return map[stat.ID]int{
			stat.Strength: 35,
			stat.Vitality: 200,
			stat.Energy:   0,
		}
	}

	if d.PlayerUnit.Stats[stat.Level] >= 30 {
		return map[stat.ID]int{
			stat.Strength:  48,
			stat.Dexterity: 40,
			stat.Vitality:  200,
			stat.Energy:    0,
		}
	}

	return map[stat.ID]int{
		stat.Strength:  0,
		stat.Dexterity: 25,
		stat.Vitality:  150,
		stat.Energy:    0,
	}
}

func (p PaladinLeveling) SkillPoints(d data.Data) []skill.Skill {
	if d.PlayerUnit.Stats[stat.Level] < 21 {
		return []skill.Skill{
			skill.Might,
			skill.Sacrifice,
			skill.ResistFire,
			skill.ResistFire,
			skill.ResistFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.Zeal,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
			skill.HolyFire,
		}
	}

	// Hammerdin
	return []skill.Skill{
		skill.HolyBolt,
		skill.BlessedHammer,
		skill.Prayer,
		skill.Defiance,
		skill.Cleansing,
		skill.Vigor,
		skill.Might,
		skill.BlessedAim,
		skill.Concentration,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		// Level 19
		skill.BlessedHammer,
		skill.Concentration,
		skill.Vigor,
		// Level 20
		skill.BlessedHammer,
		skill.Vigor,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.Vigor,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.Smite,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.BlessedHammer,
		skill.Charge,
		skill.BlessedHammer,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.HolyShield,
		skill.Concentration,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Vigor,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.Concentration,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
		skill.BlessedAim,
	}
}

func (p PaladinLeveling) GetKeyBindings(d data.Data) map[skill.Skill]string {
	skillBindings := map[skill.Skill]string{
		skill.Vigor:      config.Config.Bindings.Paladin.Vigor,
		skill.HolyShield: config.Config.Bindings.Paladin.HolyShield,
	}

	if d.PlayerUnit.Skills[skill.BlessedHammer] > 0 && d.PlayerUnit.Stats[stat.Level] >= 18 {
		skillBindings[skill.BlessedHammer] = ""
	} else if d.PlayerUnit.Skills[skill.Zeal] > 0 {
		skillBindings[skill.Zeal] = ""
	}

	if d.PlayerUnit.Skills[skill.Concentration] > 0 && d.PlayerUnit.Stats[stat.Level] >= 18 {
		skillBindings[skill.Concentration] = config.Config.Bindings.Paladin.Concentration
	} else {
		if _, found := d.PlayerUnit.Skills[skill.HolyFire]; found {
			skillBindings[skill.HolyFire] = config.Config.Bindings.Paladin.Concentration
		} else if _, found := d.PlayerUnit.Skills[skill.Might]; found {
			skillBindings[skill.Might] = config.Config.Bindings.Paladin.Concentration
		}
	}

	return skillBindings
}

func (p PaladinLeveling) ShouldResetSkills(d data.Data) bool {
	if d.PlayerUnit.Stats[stat.Level] >= 21 && d.PlayerUnit.Skills[skill.HolyFire] > 10 {
		return true
	}

	return false
}

func (p PaladinLeveling) KillAncients() action.Action {
	return action.NewChain(func(d data.Data) (actions []action.Action) {
		for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
			actions = append(actions,
				p.killMonster(m.Name, data.MonsterTypeSuperUnique),
			)
		}
		return actions
	})
}
