package character

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

const (
	paladinLevelingMaxAttacksLoop = 10
	respecLevel                   = 21
)

type PaladinLeveling struct {
	BaseCharacter
}

func (s PaladinLeveling) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s PaladinLeveling) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0

	for {
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= paladinLevelingMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		numOfAttacks := 5

		if s.Data.PlayerUnit.Skills[skill.BlessedHammer].Level > 0 {
			s.Logger.Debug("Using Blessed Hammer")
			// Add a random movement, maybe hammer is not hitting the target
			if previousUnitID == int(id) {
				if monster.Stats[stat.Life] > 0 {
					s.PathFinder.RandomMovement()
				}
				return nil
			}
			step.PrimaryAttack(id, numOfAttacks, false, step.Distance(2, 7), step.EnsureAura(skill.Concentration))

		} else {
			if s.Data.PlayerUnit.Skills[skill.Zeal].Level > 0 {
				s.Logger.Debug("Using Zeal")
				numOfAttacks = 1
			}
			s.Logger.Debug("Using primary attack with Holy Fire aura")
			step.PrimaryAttack(id, numOfAttacks, false, step.Distance(1, 3), step.EnsureAura(skill.HolyFire))
		}

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s PaladinLeveling) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s PaladinLeveling) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		skillsList = append(skillsList, skill.HolyShield)
	}
	s.Logger.Info("Buff skills", "skills", skillsList)
	return skillsList
}

func (s PaladinLeveling) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s PaladinLeveling) ShouldResetSkills() bool {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	if lvl.Value >= respecLevel && s.Data.PlayerUnit.Skills[skill.HolyFire].Level > 10 {
		s.Logger.Info(fmt.Sprintf("Resetting skills: Level %d+ and Holy Fire level > 10", respecLevel))
		return true
	}

	return false
}

func (s PaladinLeveling) SkillsToBind() (skill.ID, []skill.ID) {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	mainSkill := skill.AttackSkill
	skillBindings := []skill.ID{}

	if lvl.Value >= 6 {
		skillBindings = append(skillBindings, skill.Vigor)
	}

	if lvl.Value >= 24 {
		skillBindings = append(skillBindings, skill.HolyShield)
	}

	if s.Data.PlayerUnit.Skills[skill.BlessedHammer].Level > 0 && lvl.Value >= 18 {
		mainSkill = skill.BlessedHammer
	} else if s.Data.PlayerUnit.Skills[skill.Zeal].Level > 0 {
		mainSkill = skill.Zeal
	}

	if s.Data.PlayerUnit.Skills[skill.Concentration].Level > 0 && lvl.Value >= 18 {
		skillBindings = append(skillBindings, skill.Concentration)
	} else {
		if _, found := s.Data.PlayerUnit.Skills[skill.HolyFire]; found {
			skillBindings = append(skillBindings, skill.HolyFire)
		} else if _, found := s.Data.PlayerUnit.Skills[skill.Might]; found {
			skillBindings = append(skillBindings, skill.Might)
		}
	}

	s.Logger.Info("Skills bound", "mainSkill", mainSkill, "skillBindings", skillBindings)
	return mainSkill, skillBindings
}

func (s PaladinLeveling) StatPoints() map[stat.ID]int {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	statPoints := make(map[stat.ID]int)

	if lvl.Value <= 6 {
		statPoints[stat.Vitality] = 9999
	} else if lvl.Value < respecLevel {
		statPoints[stat.Strength] = 0
		statPoints[stat.Dexterity] = 25
		statPoints[stat.Vitality] = 150
		statPoints[stat.Energy] = 0
	} else if lvl.Value < 30 {
		statPoints[stat.Strength] = 25
		statPoints[stat.Vitality] = 210
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

	s.Logger.Info("Assigning stat points", "level", lvl.Value, "statPoints", statPoints)
	return statPoints
}

func (s PaladinLeveling) SkillPoints() []skill.ID {
	lvl, _ := s.Data.PlayerUnit.FindStat(stat.Level, 0)
	var skillPoints []skill.ID

	if lvl.Value < respecLevel {
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

	s.Logger.Info("Assigning skill points", "level", lvl.Value, "skillPoints", skillPoints)
	return skillPoints
}

func (s PaladinLeveling) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s PaladinLeveling) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeUnique)
}

func (s PaladinLeveling) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeUnique)
}

func (s PaladinLeveling) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeUnique)
}

func (s PaladinLeveling) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := s.PathFinder.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.PathFinder.DistanceFromMe(councilMembers[j].Position)

			return distanceI < distanceJ
		})

		if len(councilMembers) > 0 {
			s.Logger.Debug("Targeting Council member", "id", councilMembers[0].UnitID)
			return councilMembers[0].UnitID, true
		}

		s.Logger.Debug("No Council members found")
		return 0, false
	}, nil)
}

func (s PaladinLeveling) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeUnique)
}
func (s PaladinLeveling) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s PaladinLeveling) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			time.Sleep(200)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")

		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s PaladinLeveling) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s PaladinLeveling) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s PaladinLeveling) KillAncients() error {
	for _, m := range s.Data.Monsters.Enemies(data.MonsterEliteFilter()) {
		m, _ := s.Data.Monsters.FindOne(m.Name, data.MonsterTypeSuperUnique)

		s.killMonster(m.Name, data.MonsterTypeSuperUnique)
	}
	return nil
}

func (s PaladinLeveling) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
