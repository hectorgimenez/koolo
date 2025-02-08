package character

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	maxJavazonAttackLoops = 10
	minJavazonDistance    = 10
	maxJavazonDistance    = 30
)

type Javazon struct {
	CharacterBuild
}

func (s Javazon) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.LightningFury, skill.TomeOfTownPortal}
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

func (s Javazon) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	const numOfAttacks = 5

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

		if completedAttackLoops >= maxJavazonAttackLoops {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		closeMonsters := 0
		for _, mob := range s.Data.Monsters {
			if mob.IsPet() || mob.IsMerc() || mob.IsGoodNPC() || mob.IsSkip() || monster.Stats[stat.Life] <= 0 && mob.UnitID != monster.UnitID {
				continue
			}
			if pather.DistanceFromPoint(mob.Position, monster.Position) <= 15 {
				closeMonsters++
			}
			if closeMonsters >= 3 {
				break
			}
		}

		if closeMonsters >= 3 {
			step.SecondaryAttack(skill.LightningFury, id, numOfAttacks, step.Distance(minJavazonDistance, maxJavazonDistance))
		} else {
			step.PrimaryAttack(id, numOfAttacks, false, step.Distance(1, 1))
		}

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s Javazon) KillBossSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	const numOfAttacks = 5

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

		if completedAttackLoops >= maxJavazonAttackLoops {
			return nil
		}

		completedAttackLoops++
		previousUnitID = int(id)

		step.PrimaryAttack(id, numOfAttacks, false, step.Distance(1, 1))
	}
}

func (s Javazon) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s Javazon) killBoss(npc npc.ID, t data.MonsterType) error {
	return s.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s Javazon) PreCTABuffSkills() []skill.ID {
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Valkyrie); found {
		return []skill.ID{skill.Valkyrie}
	} else {
		return []skill.ID{}
	}
}

func (s Javazon) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
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

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}
