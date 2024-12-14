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
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	maxAttacksLoop = 5
	minDistance    = 25
	maxDistance    = 30
)

type Trapsin struct {
	BaseCharacter
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

		if !s.preBattleChecks(id) {
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
	})
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

func (s Trapsin) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Trapsin) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeUnique)
}

func (s Trapsin) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeUnique)
}

func (s Trapsin) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeUnique)
}

func (s Trapsin) KillCouncil() error {
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
	})
}

func (s Trapsin) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeUnique)
}

func (s Trapsin) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s Trapsin) KillDiablo() error {
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

func (s Trapsin) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Trapsin) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Trapsin) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
