package character

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

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
	BaseCharacter
}

func (s Javazon) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.LightningFury, skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
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
		id, found := monsterSelector(*s.data)
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

		monster, found := s.data.Monsters.FindByID(id)
		if !found {
			s.logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		closeMonsters := 0
		for _, mob := range s.data.Monsters {
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
		id, found := monsterSelector(*s.data)
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
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Valkyrie); found {
		return []skill.ID{skill.Valkyrie}
	} else {
		return []skill.ID{}
	}
}

func (s Javazon) BuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Javazon) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Javazon) KillAndariel() error {
	return s.killBoss(npc.Andariel, data.MonsterTypeNone)
}

func (s Javazon) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s Javazon) KillDuriel() error {
	return s.killBoss(npc.Duriel, data.MonsterTypeNone)
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
			distanceI := s.pf.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.pf.DistanceFromMe(councilMembers[j].Position)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s Javazon) KillMephisto() error {
	return s.killBoss(npc.Mephisto, data.MonsterTypeNone)
}

func (s Javazon) KillIzual() error {
	return s.killBoss(npc.Izual, data.MonsterTypeNone)
}

func (s Javazon) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.data.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
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
		s.logger.Info("Diablo detected, attacking")

		return s.killMonster(npc.Diablo, data.MonsterTypeNone)
	}
}

func (s Javazon) KillPindle() error {
	return s.killBoss(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Javazon) KillNihlathak() error {
	return s.killBoss(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Javazon) KillBaal() error {
	return s.killBoss(npc.BaalCrab, data.MonsterTypeNone)
}
