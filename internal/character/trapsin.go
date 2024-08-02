package character

import (
	"log/slog"
	"sort"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
)

const (
	maxAttacksLoop = 5
	minDistance    = 25
	maxDistance    = 30
)

type Trapsin struct {
	BaseCharacter
}

func (s Trapsin) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.DeathSentry, skill.LightningSentry, skill.TomeOfTownPortal}
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

func (s Trapsin) BuffSkills(d game.Data) []skill.ID {
	armor := skill.Fade
	armors := []skill.ID{skill.BurstOfSpeed, skill.Fade}
	for _, arm := range armors {
		if _, found := d.KeyBindings.KeyBindingForSkill(arm); found {
			armor = arm
		}
	}

	if _, found := d.KeyBindings.KeyBindingForSkill(skill.BladeShield); found {
		return []skill.ID{armor, skill.BladeShield}
	}

	return []skill.ID{armor}
}

func (s Trapsin) PreCTABuffSkills(d game.Data) []skill.ID {
	armor := skill.ShadowWarrior
	armors := []skill.ID{skill.ShadowWarrior, skill.ShadowMaster}
	hasShadow := false
	for _, arm := range armors {
		if _, found := d.KeyBindings.KeyBindingForSkill(arm); found {
			armor = arm
			hasShadow = true
		}
	}

	if hasShadow {
		return []skill.ID{armor}
	}

	return []skill.ID{}
}

func (s Trapsin) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if !s.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		opts := []step.AttackOption{step.Distance(minDistance, maxDistance)}

		helper.Sleep(100)
		steps = append(steps,
			step.SecondaryAttack(skill.LightningSentry, id, 3, opts...),
			step.SecondaryAttack(skill.DeathSentry, id, 2, opts...),
			step.PrimaryAttack(id, 2, true, step.Distance(minDistance, maxDistance)),
		)

		return
	}, action.RepeatUntilNoSteps())
}

func (s Trapsin) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}

		opts := []step.AttackOption{step.Distance(minDistance, maxDistance)}

		helper.Sleep(100)
		steps = append(steps,
			step.SecondaryAttack(skill.LightningSentry, m.UnitID, 3, opts...),
			step.SecondaryAttack(skill.DeathSentry, m.UnitID, 2, opts...),
			step.PrimaryAttack(m.UnitID, 2, true, opts...),
		)

		return
	}, action.CanBeSkipped())
}

func (s Trapsin) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Trapsin) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s Trapsin) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s Trapsin) KillDuriel() action.Action {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s Trapsin) KillPindle(_ []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Trapsin) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s Trapsin) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Trapsin) KillDiablo() action.Action {
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
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s Trapsin) KillIzual() action.Action {
	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s Trapsin) KillCouncil() action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		// Exclude monsters that are not council members
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

		for _, m := range councilMembers {
			for range maxAttacksLoop {
				steps = append(steps,
					step.SecondaryAttack(skill.LightningSentry, m.UnitID, 3, step.Distance(minDistance, maxDistance)),
					step.SecondaryAttack(skill.DeathSentry, m.UnitID, 2, step.Distance(minDistance, maxDistance)),
					step.PrimaryAttack(m.UnitID, 2, true, step.Distance(minDistance, maxDistance)),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s Trapsin) KillBaal() action.Action {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
