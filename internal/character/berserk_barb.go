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
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

type Berserker struct {
	BaseCharacter
	*game.HID
}

func (s Berserker) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.BattleCommand, skill.BattleOrders, skill.Shout, skill.FindItem, skill.TomeOfTownPortal}
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

func (s Berserker) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0
	findItemAttempts := 0
	const maxFindItemAttempts = 1
	const maxRange = 15
	const berserkerMaxAttacksLoop = 5 // Adjust this value as needed

	for {
		id, found := monsterSelector(*s.data)
		if !found {
			// Find Item logic
			if findItemAttempts < maxFindItemAttempts {
				foundCorpse := s.FindItemOnNearbyCorpses(*s.data, maxRange, time.Millisecond*100)
				if foundCorpse {
					findItemAttempts++
					utils.Sleep(1000)
				} else {
					findItemAttempts = maxFindItemAttempts
				}
			}

			if findItemAttempts >= maxFindItemAttempts {
				return nil
			}
			continue
		}

		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= berserkerMaxAttacksLoop {
			return nil
		}

		monster, found := s.data.Monsters.FindByID(id)
		if !found {
			s.logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		step.MoveTo(monster.Position)
		step.PrimaryAttack(id, 1, false, step.Distance(1, maxRange))

		completedAttackLoops++
		previousUnitID = int(id)
	}
}

func (s Berserker) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s *Berserker) FindItemOnNearbyCorpses(d game.Data, maxRange int, waitTime time.Duration) bool {
	s.logger.Debug("Attempting Find Item on nearby corpses", slog.Int("total_corpses", len(d.Corpses)))

	findItemKey, found := d.KeyBindings.KeyBindingForSkill(skill.FindItem)
	if !found {
		s.logger.Debug("Find Item skill not found in key bindings")
		return false
	}

	playerPos := d.PlayerUnit.Position
	corpseFound := false
	successfulFindItems := 0
	checkedCorpses := make(map[data.UnitID]bool)

	// Sort corpses by distance from the player
	sort.Slice(d.Corpses, func(i, j int) bool {
		distI := s.pf.DistanceFromMe(d.Corpses[i].Position)
		distJ := s.pf.DistanceFromMe(d.Corpses[j].Position)
		return distI < distJ
	})

	for _, corpse := range d.Corpses {
		distance := s.pf.DistanceFromMe(corpse.Position)

		if distance > maxRange {
			break
		}

		// Check if this corpse has already been processed
		if checkedCorpses[corpse.UnitID] {
			continue
		}

		if corpse.Type != data.MonsterTypeChampion &&
			corpse.Type != data.MonsterTypeMinion &&
			corpse.Type != data.MonsterTypeUnique &&
			corpse.Type != data.MonsterTypeSuperUnique {
			continue
		}

		screenX, screenY := s.pf.GameCoordsToScreenCords(
			corpse.Position.X, corpse.Position.Y,
		)

		s.HID.MovePointer(screenX, screenY)
		time.Sleep(waitTime)

		s.HID.PressKeyBinding(findItemKey)
		time.Sleep(waitTime)

		s.HID.Click(game.RightButton, screenX, screenY)
		s.logger.Debug("Find Item used on corpse", slog.Any("corpse_id", corpse.UnitID))
		time.Sleep(waitTime)

		// Mark this corpse as checked
		checkedCorpses[corpse.UnitID] = true

		corpseFound = true
		successfulFindItems++

		if d.PlayerUnit.States.HasState(state.Cooldown) {
			break
		}

		playerPos = d.PlayerUnit.Position

		if s.pf.DistanceFromMe(playerPos) > maxRange {
			break
		}
	}

	s.logger.Debug("Find Item sequence completed",
		slog.Bool("corpse_found", corpseFound),
		slog.Int("successful_find_items", successfulFindItems))
	return corpseFound
}

// Placeholder for possible Howl addition
///
// func (s Berserker) ensureHowlSkill(d game.Data) []step.Step {
// 	if d.PlayerUnit.RightSkill != skill.Howl {
// 		if howlKey, found := d.KeyBindings.KeyBindingForSkill(skill.Howl); found {
// 			s.logger.Debug(fmt.Sprintf("Activating Howl skill with key binding: %v", howlKey))
// 			return []step.Step{step.SyncStep(func(d game.Data) error {
// 				s.container.HID.PressKeyBinding(howlKey)
// 				time.Sleep(80 * time.Millisecond) // Short delay after key press
// 				return nil
// 			})}
// 		} else {
// 			s.logger.Debug("Howl key binding not found")
// 		}
// 	} else {
// 		s.logger.Debug("Howl skill already active")
// 	}
// 	return nil
// }

func (s Berserker) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Berserker) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.BattleCommand); found {
		skillsList = append(skillsList, skill.BattleCommand)
	}

	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Shout); found {
		skillsList = append(skillsList, skill.Shout)
	}

	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.BattleOrders); found {
		skillsList = append(skillsList, skill.BattleOrders)
	}

	return skillsList
}

func (s Berserker) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Berserker) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s Berserker) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s Berserker) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s Berserker) KillCouncil() error {
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

func (s Berserker) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s Berserker) KillIzual() error {
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)

	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s Berserker) KillDiablo() error {
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

func (s Berserker) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Berserker) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Berserker) KillBaal() error {
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)

	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
