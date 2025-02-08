package character

import (
	"fmt"
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxAttacksLoop = 5
	minDistance    = 25
	maxDistance    = 30
)

type Trapsin struct {
	CharacterBuild
}

func (s Trapsin) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.DeathSentry, skill.LightningSentry, skill.TomeOfTownPortal}
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

func (s Trapsin) KillMonsterSequence(
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

		if completedAttackLoops >= maxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		opts := step.Distance(minDistance, maxDistance)

		utils.Sleep(100)
		step.SecondaryAttack(skill.LightningSentry, id, 3, opts)
		step.SecondaryAttack(skill.DeathSentry, id, 2, opts)
		step.PrimaryAttack(id, 2, true, opts)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s Trapsin) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s Trapsin) BuffSkills() []skill.ID {
	armor := skill.Fade
	armors := []skill.ID{skill.BurstOfSpeed, skill.Fade}
	for _, arm := range armors {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(arm); found {
			armor = arm
		}
	}

	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.BladeShield); found {
		return []skill.ID{armor, skill.BladeShield}
	}

	return []skill.ID{armor}
}

func (s Trapsin) PreCTABuffSkills() []skill.ID {
	armor := skill.ShadowWarrior
	armors := []skill.ID{skill.ShadowWarrior, skill.ShadowMaster}
	hasShadow := false
	for _, arm := range armors {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(arm); found {
			armor = arm
			hasShadow = true
		}
	}

	if hasShadow {
		return []skill.ID{armor}
	}

	return []skill.ID{}
}
