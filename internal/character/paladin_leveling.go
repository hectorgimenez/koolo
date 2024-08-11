package character

import (
	"sort"
	"time"

	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/pather"
)

type PaladinLeveling struct {
	BaseCharacter
}

func (p PaladinLeveling) CheckKeyBindings(d game.Data) []skill.ID {

	// Not implemented
	return []skill.ID{}
}

func (p PaladinLeveling) BuffSkills(d game.Data) []skill.ID {
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return []skill.ID{}
}

func (p PaladinLeveling) PreCTABuffSkills(_ game.Data) []skill.ID {
	return []skill.ID{}
}

func (p PaladinLeveling) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return p.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
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
	return p.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
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
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			p.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			return []action.Action{action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			})}
		}

		diabloFound = true
		p.logger.Info("Diablo detected, attacking")

		return []action.Action{
			p.killMonster(npc.Diablo, data.MonsterTypeNone),
			p.killMonster(npc.Diablo, data.MonsterTypeNone),
			p.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (p PaladinLeveling) KillIzual() action.Action {
	return p.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillBaal() action.Action {
	return p.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
	completedAttackLoops := 0
	var previousUnitID data.UnitID = 0

	return action.NewStepChain(func(d game.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if previousUnitID != id {
			completedAttackLoops = 0
		}

		if !p.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		if completedAttackLoops >= 10 {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		numOfAttacks := 5

		if d.PlayerUnit.Skills[skill.BlessedHammer].Level > 0 {
			// Add a random movement, maybe hammer is not hitting the target
			if previousUnitID == id {
				steps = append(steps,
					step.SyncStep(func(_ game.Data) error {
						p.container.PathFinder.RandomMovement(d)
						return nil
					}),
				)
			}
			steps = append(steps,
				step.PrimaryAttack(id, numOfAttacks, false, step.Distance(2, 7), step.EnsureAura(skill.Concentration)),
			)
		} else {
			if d.PlayerUnit.Skills[skill.Zeal].Level > 0 {
				numOfAttacks = 1
			}

			steps = append(steps,
				step.PrimaryAttack(id, numOfAttacks, false, step.Distance(1, 3), step.EnsureAura(skill.HolyFire)),
			)
		}

		completedAttackLoops++
		previousUnitID = id
		return steps
	}, action.RepeatUntilNoSteps())
}

func (p PaladinLeveling) StatPoints(d game.Data) map[stat.ID]int {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)

	if lvl.Value >= 21 && lvl.Value < 30 {
		return map[stat.ID]int{
			stat.Strength: 35,
			stat.Vitality: 200,
			stat.Energy:   0,
		}
	}

	if lvl.Value >= 30 && lvl.Value < 45 {
		return map[stat.ID]int{
			stat.Strength:  50,
			stat.Dexterity: 40,
			stat.Vitality:  220,
			stat.Energy:    0,
		}
	}

	if lvl.Value >= 45 {
		return map[stat.ID]int{
			stat.Strength:  86,
			stat.Dexterity: 50,
			stat.Vitality:  300,
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

func (p PaladinLeveling) SkillPoints(d game.Data) []skill.ID {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value < 21 {
		return []skill.ID{
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
	return []skill.ID{
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

func (p PaladinLeveling) SkillsToBind(d game.Data) (skill.ID, []skill.ID) {
	mainSkill := skill.AttackSkill
	skillBindings := []skill.ID{
		skill.Vigor,
		skill.HolyShield,
	}

	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if d.PlayerUnit.Skills[skill.BlessedHammer].Level > 0 && lvl.Value >= 18 {
		mainSkill = skill.BlessedHammer
	} else if d.PlayerUnit.Skills[skill.Zeal].Level > 0 {
		mainSkill = skill.Zeal
	}

	if d.PlayerUnit.Skills[skill.Concentration].Level > 0 && lvl.Value >= 18 {
		skillBindings = append(skillBindings, skill.Concentration)
	} else {
		if _, found := d.PlayerUnit.Skills[skill.HolyFire]; found {
			skillBindings = append(skillBindings, skill.HolyFire)
		} else if _, found := d.PlayerUnit.Skills[skill.Might]; found {
			skillBindings = append(skillBindings, skill.Might)
		}
	}

	return mainSkill, skillBindings
}

func (p PaladinLeveling) ShouldResetSkills(d game.Data) bool {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 21 && d.PlayerUnit.Skills[skill.HolyFire].Level > 10 {
		return true
	}

	return false
}

func (p PaladinLeveling) KillAncients() action.Action {
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
			actions = append(actions,
				p.killMonster(m.Name, data.MonsterTypeSuperUnique),
			)
		}
		return actions
	})
}
