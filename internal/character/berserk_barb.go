package character

import (
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/d2go/pkg/utils"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Berserker struct {
	BaseCharacter
}

func (s Berserker) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.BattleCommand, skill.BattleOrders, skill.Shout, skill.FindItem, skill.TomeOfTownPortal}
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

func (s Berserker) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	var attackAttempts int
	var findItemAttempts int
	const maxFindItemAttempts = 1
	const maxRange = 15

	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		// Monster Selection
		id, found := monsterSelector(d)
		if found {
			s.logger.Debug("Monster found", slog.Int("unitID", int(id)))
			attackAttempts = 0

			monster, monsterFound := d.Monsters.FindByID(id)
			if !monsterFound {
				s.logger.Debug("Monster data not found")
				return []step.Step{}
			}

			if !s.preBattleChecks(d, id, skipOnImmunities) {
				s.logger.Debug("Pre-battle checks failed")
				return []step.Step{}
			}

			steps = append(steps, step.MoveTo(monster.Position))
			s.logger.Debug("Added step to move towards monster", slog.Any("position", monster.Position))

			s.logger.Debug("Adding primary attack step", slog.Int("attemptNumber", attackAttempts+1))
			attackStep := step.PrimaryAttack(id, 1, false, step.Distance(1, maxRange))
			steps = append(steps, step.SyncStep(func(d game.Data) error {
				s.logger.Debug("Executing primary attack", slog.Int("attemptNumber", attackAttempts+1))
				err := attackStep.Run(d, s.container)
				if err != nil {
					s.logger.Debug("Primary attack failed", slog.String("error", err.Error()))
				} else {
					s.logger.Debug("Primary attack executed successfully")
				}
				attackAttempts++
				return err
			}))

			if attackAttempts >= 5 {
				s.logger.Debug("Resetting attack sequence after multiple attempts")
				attackAttempts = 0
			}
		} else {
			s.logger.Debug("No targetable monster found")
			attackAttempts = 0

			// If no close monsters, attempt Find Item
			if findItemAttempts < maxFindItemAttempts {
				foundCorpse := s.FindItemOnNearbyCorpses(d, maxRange, time.Millisecond*100)
				if foundCorpse {
					findItemAttempts++
					steps = append(steps, step.Wait(time.Millisecond*100))
				} else {
					findItemAttempts = maxFindItemAttempts // Ensure we don't try Find Item again if no corpses found
				}
			}

			if findItemAttempts >= maxFindItemAttempts {
				s.logger.Debug("Find Item attempts exhausted, exiting sequence")
				return nil
			}
		}

		s.logger.Debug("Steps created", slog.Int("stepCount", len(steps)))
		return steps
	}, action.RepeatUntilNoSteps())
}

func (s *Berserker) FindItemOnNearbyCorpses(d game.Data, maxRange int, waitTime time.Duration) bool {
	s.logger.Debug("Attempting Find Item on nearby corpses", slog.Int("total_corpses", len(d.Corpses)))

	findItemKey, found := d.KeyBindings.KeyBindingForSkill(skill.FindItem)
	if !found {
		s.logger.Debug("Find Item skill not found in key bindings")
		return false
	}

	playerPos := d.PlayerUnit.Position
	originalPosition := playerPos
	corpseFound := false
	successfulFindItems := 0
	checkedCorpses := make(map[data.UnitID]bool)

	// Sort corpses by distance from the player
	sort.Slice(d.Corpses, func(i, j int) bool {
		distI := utils.DistanceFromPoint(playerPos, d.Corpses[i].Position)
		distJ := utils.DistanceFromPoint(playerPos, d.Corpses[j].Position)
		return distI < distJ
	})

	for _, corpse := range d.Corpses {
		distance := utils.DistanceFromPoint(playerPos, corpse.Position)

		if distance > maxRange {
			s.logger.Debug("Reached corpse beyond maxRange, stopping Find Item process",
				slog.Int("distance", int(distance)),
				slog.Int("maxRange", maxRange))
			break
		}

		// Check if this corpse has already been processed
		if checkedCorpses[corpse.UnitID] {
			s.logger.Debug("Skipping already checked corpse", slog.Any("corpse_id", corpse.UnitID))
			continue
		}

		if corpse.Type != data.MonsterTypeChampion &&
			corpse.Type != data.MonsterTypeMinion &&
			corpse.Type != data.MonsterTypeUnique &&
			corpse.Type != data.MonsterTypeSuperUnique {
			s.logger.Debug("Skipping corpse: not a desired monster type",
				slog.Any("corpse_id", corpse.UnitID),
				slog.Any("monster_type", corpse.Type))
			continue
		}

		s.logger.Debug("Preparing to use Find Item on corpse",
			slog.Any("corpse_id", corpse.UnitID),
			slog.Any("corpse_position", corpse.Position),
			slog.Any("monster_type", corpse.Type),
			slog.Int("distance", int(distance)))

		screenX, screenY := s.container.PathFinder.GameCoordsToScreenCords(
			playerPos.X, playerPos.Y,
			corpse.Position.X, corpse.Position.Y,
		)

		s.logger.Debug("Find Item screen coordinates",
			slog.Int("screenX", screenX),
			slog.Int("screenY", screenY),
			slog.Any("playerPos", playerPos),
			slog.Any("corpsePos", corpse.Position))

		s.container.HID.MovePointer(screenX, screenY)
		time.Sleep(waitTime)

		s.container.HID.PressKeyBinding(findItemKey)
		time.Sleep(waitTime)

		s.container.HID.Click(game.RightButton, screenX, screenY)
		s.logger.Debug("Find Item used on corpse", slog.Any("corpse_id", corpse.UnitID))
		time.Sleep(waitTime)

		// Mark this corpse as checked
		checkedCorpses[corpse.UnitID] = true

		corpseFound = true
		successfulFindItems++

		if d.PlayerUnit.States.HasState(state.Cooldown) {
			s.logger.Debug("Find Item on cooldown, stopping")
			break
		}

		playerPos = d.PlayerUnit.Position

		if utils.DistanceFromPoint(originalPosition, playerPos) > maxRange {
			s.logger.Debug("Moved too far from original position, stopping Find Item",
				slog.Any("originalPosition", originalPosition),
				slog.Any("currentPosition", playerPos),
				slog.Int("distance", int(utils.DistanceFromPoint(originalPosition, playerPos))))
			break
		}
	}

	s.logger.Debug("Find Item sequence completed",
		slog.Bool("corpse_found", corpseFound),
		slog.Int("successful_find_items", successfulFindItems))
	return corpseFound
}

func (s Berserker) PreCTABuffSkills(d game.Data) []skill.ID {
	return []skill.ID{}
}

func (s Berserker) BuffSkills(d game.Data) []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.BattleCommand); found {
		skillsList = append(skillsList, skill.BattleCommand)
	}

	if _, found := d.KeyBindings.KeyBindingForSkill(skill.Shout); found {
		skillsList = append(skillsList, skill.Shout)
	}

	if _, found := d.KeyBindings.KeyBindingForSkill(skill.BattleOrders); found {
		skillsList = append(skillsList, skill.BattleOrders)
	}

	return skillsList
}

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

func (s Berserker) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {

		monster, found := d.Monsters.FindOne(npc, t)
		if !found {
			return []step.Step{}
		}

		opts := []step.AttackOption{step.Distance(1, 2)}

		// Finish it off with primary attack
		steps = append(steps, step.PrimaryAttack(monster.UnitID, 1, false, opts...))

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s Berserker) KillCouncil() action.Action {
	var councilKilled bool
	var findItemAttempts int
	const maxFindItemAttempts = 1
	const maxRange = 30

	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// If council is killed and we've done enough Find Item attempts, exit the sequence
		if councilKilled && findItemAttempts >= maxFindItemAttempts {
			return 0, false
		}

		// If council is killed, attempt Find Item
		if councilKilled {
			findItemAttempts++
			corpseFound := s.FindItemOnNearbyCorpses(d, maxRange, time.Millisecond*100)
			if !corpseFound {
				return 0, false
			}
			// If Find Item was successful, continue the sequence to potentially find more council members
			councilKilled = false
			return 0, true
		}

		// Logic to find and target council members
		for _, mobs := range d.Monsters.Enemies() {
			if (mobs.Name == npc.CouncilMember || mobs.Name == npc.CouncilMember2 || mobs.Name == npc.CouncilMember3) &&
				utils.DistanceFromPoint(d.PlayerUnit.Position, mobs.Position) <= maxRange {
				return mobs.UnitID, true
			}
		}

		// If no council members found, mark as killed and attempt Find Item on next iteration
		councilKilled = true
		return 0, true
	}, nil, step.Distance(1, maxRange))
}

func (s Berserker) KillBaal() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
			s.killMonster(npc.BaalCrab, data.MonsterTypeNone),
		}
	})
}

func (s Berserker) KillIzual() action.Action {
	return action.NewChain(func(d game.Data) []action.Action {
		return []action.Action{
			// We will need a lot of cycles to kill him probably
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
			s.killMonster(npc.Izual, data.MonsterTypeNone),
		}
	})
}

func (s Berserker) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
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
		s.logger.Info("Diablo detected, attacking")

		return []action.Action{
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s Berserker) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Berserker) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s Berserker) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Berserker) KillDuriel() action.Action {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s Berserker) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s Berserker) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s Berserker) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}
