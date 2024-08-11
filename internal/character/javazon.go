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

func (s Javazon) CheckKeyBindings(d game.Data) []skill.ID {
	requireKeybindings := []skill.ID{skill.LightningFury, skill.TomeOfTownPortal}
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

func (s Javazon) PreCTABuffSkills(d game.Data) []skill.ID {
	if _, found := d.KeyBindings.KeyBindingForSkill(skill.Valkyrie); found {
		return []skill.ID{skill.Valkyrie}
	} else {
		return []skill.ID{}
	}
}

func (s Javazon) BuffSkills(d game.Data) []skill.ID {
	return []skill.ID{}
}

func (a Javazon) KillBossSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
	completedAttackLoops := 0
	var previousUnitID data.UnitID = 0
	skipOnImmunities = append(skipOnImmunities, stat.LightImmune)

	return action.NewStepChain(func(d game.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if previousUnitID != id {
			completedAttackLoops = 0
		}

		if !a.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		if completedAttackLoops >= maxJavazonAttackLoops {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		numOfAttacks := 5

		steps = append(steps, step.PrimaryAttack(id, numOfAttacks, false, step.Distance(1, 1)))

		completedAttackLoops++
		previousUnitID = id
		return steps
	}, action.RepeatUntilNoSteps())
}

func (a Javazon) KillMonsterSequence(monsterSelector func(d game.Data) (data.UnitID, bool), skipOnImmunities []stat.Resist, opts ...step.AttackOption) action.Action {
	completedAttackLoops := 0
	var previousUnitID data.UnitID = 0

	//skipOnImmunities = append(skipOnImmunities, stat.LightImmune)

	return action.NewStepChain(func(d game.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if previousUnitID != id {
			completedAttackLoops = 0
		}

		if !a.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		if completedAttackLoops >= 10 {
			return []step.Step{}
		}

		steps := make([]step.Step, 0)

		numOfAttacks := 5

		monster, found := d.Monsters.FindByID(id)
		if !found {
			return []step.Step{}
		}

		closeMonsters := 0
		for _, mob := range d.Monsters {
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
			steps = append(steps, step.SecondaryAttack(skill.LightningFury, id, numOfAttacks, step.Distance(minJavazonDistance, maxJavazonDistance)))
		} else {
			steps = append(steps, step.PrimaryAttack(id, numOfAttacks, false, step.Distance(1, 1)))
		}

		completedAttackLoops++
		previousUnitID = id
		return steps
	}, action.RepeatUntilNoSteps())
}

func (a Javazon) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return a.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)

		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (a Javazon) killBoss(npc npc.ID, t data.MonsterType) action.Action {
	return a.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)

		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (a Javazon) KillCountess() action.Action {
	return a.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (a Javazon) KillAndariel() action.Action {
	return a.killBoss(npc.Andariel, data.MonsterTypeNone)
}

func (a Javazon) KillSummoner() action.Action {
	return a.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (a Javazon) KillDuriel() action.Action {
	return a.killBoss(npc.Duriel, data.MonsterTypeNone)
}

func (a Javazon) KillMephisto() action.Action {
	return a.killBoss(npc.Mephisto, data.MonsterTypeNone)
}

func (a Javazon) KillPindle(_ []stat.Resist) action.Action {
	return a.killBoss(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (a Javazon) KillNihlathak() action.Action {
	return a.killBoss(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (a Javazon) KillCouncil() action.Action {
	return a.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
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

func (a Javazon) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			a.logger.Error("Diablo was not found, timeout reached")
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
		a.logger.Info("Diablo detected, attacking")

		return []action.Action{
			a.killBoss(npc.Diablo, data.MonsterTypeNone),
			a.killBoss(npc.Diablo, data.MonsterTypeNone),
			a.killBoss(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (a Javazon) KillIzual() action.Action {
	return a.killBoss(npc.Izual, data.MonsterTypeNone)
}

func (a Javazon) KillBaal() action.Action {
	return a.killBoss(npc.BaalCrab, data.MonsterTypeNone)
}
