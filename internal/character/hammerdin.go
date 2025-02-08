package character

import (
	"fmt"
	"log/slog"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

const (
	hammerdinMaxAttacksLoop = 20 // Adjust from 5-20 depending on DMG and rotation, lower attack loops would cause higher attack rotation whereas bigger would perform multiple(longer) attacks on one spot.
)

type Hammerdin struct {
	CharacterBuild
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

		if completedAttackLoops >= hammerdinMaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// Add a random movement, maybe hammer is not hitting the target
		if previousUnitID == int(id) {
			if monster.Stats[stat.Life] > 0 {
				s.PathFinder.RandomMovement()
			}
			return nil
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

func (s Hammerdin) BuffSkills() []skill.ID {
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.HolyShield); found {
		return []skill.ID{skill.HolyShield}
	}
	return []skill.ID{}
}
