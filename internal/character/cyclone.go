package character

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/utils"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
	"log/slog"
	"sort"
	"time"
)

type CycloneBarb struct {
	BaseCharacter
}

const (
	maxWWDistance     = 10
	minWWDistance     = 2
	maxWWAttackCycles = 10
	wwDistanceOffset  = 10
	maxHorkRange      = 20
)

func (c CycloneBarb) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.Whirlwind, skill.TomeOfTownPortal}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := d.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	if len(missingKeybindings) > 0 {
		c.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (c CycloneBarb) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	var previousUnitID data.UnitID
	checkedCorpses := make([]data.Monster, 0)
	attackCount := 0
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {

		id, found := monsterSelector(d)
		if !found {
			if _, findItemFound := d.KeyBindings.KeyBindingForSkill(skill.FindItem); findItemFound {
				c.logger.Debug("Find Item hotkey found, attempting to use")
				corpseFound := c.FindNearbyCorpse(d, maxHorkRange, checkedCorpses)
				if corpseFound != nil {
					checkedCorpses = append(checkedCorpses, *corpseFound)
					c.FindItem(*corpseFound, d)
					return append(steps, step.Wait(time.Millisecond*100))
				}
			}

			return steps
		}

		if !c.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		// Reset variables if new monster
		if previousUnitID != id {
			attackCount = 0
			previousUnitID = id
			c.logger.Debug("New target detected", slog.Int("unitID", int(id)))
		}

		if attackCount >= maxWWAttackCycles {
			return []step.Step{}
		}

		// Check if we have the required states (charges)
		opts = []step.AttackOption{step.Distance(1, 4)}

		opts = append(opts, step.WithDistanceOffset(wwDistanceOffset))
		opts = append(opts, step.Distance(minWWDistance, maxWWDistance))

		steps = append(steps, step.SecondaryAttack(skill.Whirlwind, id, 5, opts...))

		attackCount++

		return steps
	}, action.RepeatUntilNoSteps())
}

func (c CycloneBarb) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}

		var opts []step.AttackOption
		opts = append(opts, step.WithDistanceOffset(10))
		opts = append(opts, step.Distance(minWWDistance, maxWWDistance))

		helper.Sleep(100)
		for range maxWWAttackCycles {
			steps = append(steps, step.SecondaryAttack(skill.Whirlwind, m.UnitID, 5, opts...))
		}

		return
	}, action.CanBeSkipped())
}

func (c CycloneBarb) BuffSkills(d game.Data) []skill.ID {
	return []skill.ID{}
}

func (c CycloneBarb) PreCTABuffSkills(d game.Data) []skill.ID {
	return []skill.ID{skill.BattleCommand, skill.BattleOrders, skill.Shout}
}

func (c CycloneBarb) FindNearbyCorpse(d game.Data, maxRange int, excludedCorpses []data.Monster) *data.Monster {
	playerPos := d.PlayerUnit.Position
	sort.Slice(d.Corpses, func(i, j int) bool {
		distI := utils.DistanceFromPoint(playerPos, d.Corpses[i].Position)
		distJ := utils.DistanceFromPoint(playerPos, d.Corpses[j].Position)
		return distI < distJ
	})

	for _, corpse := range d.Corpses {
		excluded := false
		for _, excludedMob := range excludedCorpses {
			if excludedMob.UnitID == corpse.UnitID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		distance := utils.DistanceFromPoint(playerPos, corpse.Position)

		if distance > maxRange {
			break
		}

		if corpse.Type != data.MonsterTypeChampion &&
			corpse.Type != data.MonsterTypeMinion &&
			corpse.Type != data.MonsterTypeUnique &&
			corpse.Type != data.MonsterTypeSuperUnique {
			continue
		}

		return &corpse
	}

	return nil
}

func (c CycloneBarb) FindItem(corpse data.Monster, d game.Data) {
	playerPos := d.PlayerUnit.Position

	screenX, screenY := c.container.PathFinder.GameCoordsToScreenCords(
		playerPos.X, playerPos.Y,
		corpse.Position.X, corpse.Position.Y,
	)

	c.container.HID.MovePointer(screenX, screenY)
	time.Sleep(time.Millisecond * 100)

	// Use the Find Item skill
	findItemKey, _ := d.KeyBindings.KeyBindingForSkill(skill.FindItem)
	c.container.HID.PressKeyBinding(findItemKey)
	time.Sleep(time.Millisecond * 100)

	c.container.HID.Click(game.RightButton, screenX, screenY)
	time.Sleep(time.Millisecond * 100)
}

func (c CycloneBarb) KillCountess() action.Action {
	return c.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (c CycloneBarb) KillAndariel() action.Action {
	return c.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (c CycloneBarb) KillSummoner() action.Action {
	return c.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (c CycloneBarb) KillDuriel() action.Action {
	return c.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (c CycloneBarb) KillPindle(_ []stat.Resist) action.Action {
	return c.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (c CycloneBarb) KillMephisto() action.Action {
	return c.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (c CycloneBarb) KillNihlathak() action.Action {
	return c.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (c CycloneBarb) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			c.logger.Error("Diablo was not found, timeout reached")
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
		c.logger.Info("Diablo detected, attacking")

		return []action.Action{
			c.killMonster(npc.Diablo, data.MonsterTypeNone),
			c.killMonster(npc.Diablo, data.MonsterTypeNone),
			c.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (c CycloneBarb) KillIzual() action.Action {
	return c.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (c CycloneBarb) KillCouncil() action.Action {
	return c.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
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
			return councilMembers[0].UnitID, true
		}

		return 0, false
	}, nil)
}

func (c CycloneBarb) KillBaal() action.Action {
	return c.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
