package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	hammerdinMaxAttacksLoop = 20 // Adjust from 5-20 depending on DMG and rotation, lower attack loops would cause higher attack rotation whereas bigger would perform multiple(longer) attacks on one spot.
)

type Hammerdin struct {
	BaseCharacter
}

func (s Hammerdin) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Concentration, skill.HolyShield, skill.TomeOfTownPortal}
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

func (s Hammerdin) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	consecutiveAttacks := 0
	ctx := context.Get()

	for {
		ctx.PauseIfNotPriority()

		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
			consecutiveAttacks = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= hammerdinMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		if previousUnitID == int(id) && monster.Stats[stat.Life] > 0 {
			consecutiveAttacks++
			if consecutiveAttacks >= 5 { //adjust if needed -> higher value = more attacks without randommovement
				s.PathFinder.RandomMovement()
				time.Sleep(200 * time.Millisecond)
				consecutiveAttacks = 0
				continue
			}
		}

		step.PrimaryAttack(
			id,
			3,
			true,
			step.Distance(2, 2), // X,Y coords of 2,2 is the perfect hammer angle attack for NPC targeting/attacking, you can adjust accordingly anything between 1,1 - 3,3 is acceptable, where the higher the number, the bigger the distance from the player (usually used for De Seis)
			step.EnsureAura(skill.Concentration),
		)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s Hammerdin) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s Hammerdin) killMonsterByName(id npc.ID, monsterType data.MonsterType) error {
	for {
		if m, found := s.Data.Monsters.FindOne(id, monsterType); found {
			if m.Stats[stat.Life] <= 0 {
				break
			}

			s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				if m, found := d.Monsters.FindOne(id, monsterType); found {
					return m.UnitID, true
				}
				return 0, false
			}, nil)
		} else {
			break
		}
	}
	return nil
}

func (s Hammerdin) BuffSkills() []skill.ID {
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return []skill.ID{}
}

func (s Hammerdin) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Hammerdin) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillAndariel() error {
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique)
}

func (s Hammerdin) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique)
}

func (s Hammerdin) KillDuriel() error {
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique)
}

func (s Hammerdin) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies() {
			if (m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3) && m.Stats[stat.Life] > 0 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}

func (s Hammerdin) KillMephisto() error {
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique)
}

func (s Hammerdin) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s Hammerdin) KillDiablo() error {
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

func (s Hammerdin) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Hammerdin) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
