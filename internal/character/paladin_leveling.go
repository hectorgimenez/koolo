package character

import (
	"log/slog"
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

func (s PaladinLeveling) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := d.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (p PaladinLeveling) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
	completedAttackLoops := 0
	var previousUnitID data.UnitID = 0

	return action.NewStepChain(func(d game.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			p.logger.Debug("No monster found to attack")
			return []step.Step{}
		}
		if previousUnitID != id {
			p.logger.Info("New monster targeted", "id", id)
			completedAttackLoops = 0
		}

		if !p.preBattleChecks(d, id, skipOnImmunities) {
			p.logger.Debug("Pre-battle checks failed")
			return []step.Step{}
		}

		if completedAttackLoops >= 10 {
			p.logger.Info("Max attack loops reached", "loops", completedAttackLoops)
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		numOfAttacks := 5

		if d.PlayerUnit.Skills[skill.BlessedHammer].Level > 0 {
			p.logger.Debug("Using Blessed Hammer")
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
				p.logger.Debug("Using Zeal")
				numOfAttacks = 1
			}

			p.logger.Debug("Using primary attack with Holy Fire aura")
			steps = append(steps,
				step.PrimaryAttack(id, numOfAttacks, false, step.Distance(1, 3), step.EnsureAura(skill.HolyFire)),
			)
		}

		completedAttackLoops++
		previousUnitID = id
		p.logger.Debug("Attack sequence completed", "steps", len(steps), "loops", completedAttackLoops)
		return steps
	}, action.RepeatUntilNoSteps())
}

func (p PaladinLeveling) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	p.logger.Info("Killing monster", "npc", npc, "type", t)
	return p.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (p PaladinLeveling) BuffSkills(d game.Data) []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		skillsList = append(skillsList, skill.HolyShield)
	}
	p.logger.Info("Buff skills", "skills", skillsList)
	return skillsList
}

func (p PaladinLeveling) PreCTABuffSkills(_ game.Data) []skill.ID {
	return []skill.ID{}
}

func (p PaladinLeveling) ShouldResetSkills(d game.Data) bool {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= 21 && d.PlayerUnit.Skills[skill.HolyFire].Level > 10 {
		p.logger.Info("Resetting skills: Level 21+ and Holy Fire level > 10")
		return true
	}

	return false
}

func (p PaladinLeveling) SkillsToBind(d game.Data) (skill.ID, []skill.ID) {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	mainSkill := skill.AttackSkill
	skillBindings := []skill.ID{}

	if lvl.Value >= 6 {
		skillBindings = append(skillBindings, skill.Vigor)
	}

	if lvl.Value >= 24 {
		skillBindings = append(skillBindings, skill.HolyShield)
	}

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

	p.logger.Info("Skills bound", "mainSkill", mainSkill, "skillBindings", skillBindings)
	return mainSkill, skillBindings
}

func (p PaladinLeveling) StatPoints(d game.Data) map[stat.ID]int {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	statPoints := make(map[stat.ID]int)

	if lvl.Value < 21 {
		statPoints[stat.Strength] = 0
		statPoints[stat.Dexterity] = 25
		statPoints[stat.Vitality] = 150
		statPoints[stat.Energy] = 0
	} else if lvl.Value < 30 {
		statPoints[stat.Strength] = 35
		statPoints[stat.Vitality] = 200
		statPoints[stat.Energy] = 0
	} else if lvl.Value < 45 {
		statPoints[stat.Strength] = 50
		statPoints[stat.Dexterity] = 40
		statPoints[stat.Vitality] = 220
		statPoints[stat.Energy] = 0
	} else {
		statPoints[stat.Strength] = 86
		statPoints[stat.Dexterity] = 50
		statPoints[stat.Vitality] = 300
		statPoints[stat.Energy] = 0
	}

	p.logger.Info("Assigning stat points", "level", lvl.Value, "statPoints", statPoints)
	return statPoints
}

func (p PaladinLeveling) SkillPoints(d game.Data) []skill.ID {
	lvl, _ := d.PlayerUnit.FindStat(stat.Level, 0)
	var skillPoints []skill.ID

	if lvl.Value < 21 {
		skillPoints = []skill.ID{
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
	} else {
		// Hammerdin
		skillPoints = []skill.ID{
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

	p.logger.Info("Assigning skill points", "level", lvl.Value, "skillPoints", skillPoints)
	return skillPoints
}

func (p PaladinLeveling) KillCountess() action.Action {
	p.logger.Info("Starting Countess kill sequence")
	return p.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (p PaladinLeveling) KillAndariel() action.Action {
	p.logger.Info("Starting Andariel kill sequence")
	return p.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillSummoner() action.Action {
	p.logger.Info("Starting Summoner kill sequence")
	return p.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillDuriel() action.Action {
	p.logger.Info("Starting Duriel kill sequence")
	return p.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillCouncil() action.Action {
	p.logger.Info("Starting Council kill sequence")
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
			p.logger.Debug("Targeting Council member", "id", councilMembers[0].UnitID)
			return councilMembers[0].UnitID, true
		}

		p.logger.Debug("No Council members found")
		return 0, false
	}, nil)
}

func (p PaladinLeveling) KillMephisto() action.Action {
	p.logger.Info("Starting Mephisto kill sequence")
	return p.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillIzual() action.Action {
	p.logger.Info("Starting Izual kill sequence")
	return p.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (p PaladinLeveling) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
			p.logger.Info("Starting Diablo kill sequence")
		}

		if time.Since(startTime) > timeout && !diabloFound {
			p.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			if diabloFound {
				p.logger.Info("Diablo killed or not found")
				return nil
			}

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

func (p PaladinLeveling) KillPindle(_ []stat.Resist) action.Action {
	p.logger.Info("Starting Pindleskin kill sequence")
	return p.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (p PaladinLeveling) KillNihlathak() action.Action {
	p.logger.Info("Starting Nihlathak kill sequence")
	return p.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (p PaladinLeveling) KillAncients() action.Action {
	p.logger.Info("Starting Ancients kill sequence")
	return action.NewChain(func(d game.Data) (actions []action.Action) {
		for _, m := range d.Monsters.Enemies(data.MonsterEliteFilter()) {
			actions = append(actions,
				action.NewStepChain(func(d game.Data) []step.Step {
					m, _ := d.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)
					p.logger.Info("Targeting Ancient", "name", m.Name)
					return []step.Step{}
				}),
				p.killMonster(m.Name, data.MonsterTypeSuperUnique),
			)
		}
		return actions
	})
}

func (p PaladinLeveling) KillBaal() action.Action {
	p.logger.Info("Starting Baal kill sequence")
	return p.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
