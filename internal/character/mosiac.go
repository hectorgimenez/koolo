package character

import (
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type MosiacSin struct {
	BaseCharacter
}

func (s MosiacSin) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {

		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}

		// Check if we have the required states (charges)
		opts := []step.AttackOption{step.Distance(1, 2)}

		if !s.MobAlive(id, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Tiger Strike
		if !d.PlayerUnit.States.HasState(state.Tigerstrike) {
			steps = append(steps, step.SecondaryAttack(skill.TigerStrike, id, 4, opts...))
		}

		if !s.MobAlive(id, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Cobra Strike
		if !d.PlayerUnit.States.HasState(state.Cobrastrike) {
			steps = append(steps, step.SecondaryAttack(skill.CobraStrike, id, 4, opts...))
		}

		if !s.MobAlive(id, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Phoenix Strike
		if !d.PlayerUnit.States.HasState(state.Phoenixstrike) {
			steps = append(steps, step.SecondaryAttack(skill.PhoenixStrike, id, 4, opts...))
		}

		if !s.MobAlive(id, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Claws of Thunder
		if !d.PlayerUnit.States.HasState(state.Clawsofthunder) {
			steps = append(steps, step.SecondaryAttack(skill.ClawsOfThunder, id, 4, opts...))
		}

		if !s.MobAlive(id, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Blades of Ice
		if !d.PlayerUnit.States.HasState(state.Bladesofice) {
			steps = append(steps, step.SecondaryAttack(skill.BladesOfIce, id, 4, opts...))
		}

		if !s.MobAlive(id, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Fists of Fire
		//if !d.PlayerUnit.States.HasState(state.Fistsoffire) {
		//	steps = append(steps, step.SecondaryAttack(skill.FistsOfFire, id, 4, opts...))
		//}

		// Finish it off with primary attack
		steps = append(steps, step.PrimaryAttack(id, 1, opts...))

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s MosiacSin) MobAlive(mob data.UnitID, d game.Data) bool {
	_, mFround := d.Monsters.FindByID(mob)
	return mFround
}

func (s MosiacSin) BuffSkills(d game.Data) []skill.ID {
	skillsList := make([]skill.ID, 0)

	if _, found := d.KeyBindings.KeyBindingForSkill(skill.Fade); found {
		skillsList = append(skillsList, skill.Fade)
	} else {

		// If we don't use fade but we use Burst of Speed
		if _, found := d.KeyBindings.KeyBindingForSkill(skill.BurstOfSpeed); found {
			skillsList = append(skillsList, skill.BurstOfSpeed)
		}
	}

	return skillsList
}

func (s MosiacSin) PreCTABuffSkills(d game.Data) []skill.ID {
	return []skill.ID{}
}

func (s MosiacSin) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {

		monster, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}

		// Check if we have the required states (charges)
		opts := []step.AttackOption{step.Distance(1, 2)}

		// Tiger Strike
		if !d.PlayerUnit.States.HasState(state.Tigerstrike) {
			steps = append(steps, step.SecondaryAttack(skill.TigerStrike, monster.UnitID, 4, opts...))
		}

		if !s.MobAlive(monster.UnitID, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Cobra Strike
		if !d.PlayerUnit.States.HasState(state.Cobrastrike) {
			steps = append(steps, step.SecondaryAttack(skill.CobraStrike, monster.UnitID, 4, opts...))
		}

		if !s.MobAlive(monster.UnitID, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Phoenix Strike
		if !d.PlayerUnit.States.HasState(state.Phoenixstrike) {
			steps = append(steps, step.SecondaryAttack(skill.PhoenixStrike, monster.UnitID, 4, opts...))
		}

		if !s.MobAlive(monster.UnitID, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Claws of Thunder
		if !d.PlayerUnit.States.HasState(state.Clawsofthunder) {
			steps = append(steps, step.SecondaryAttack(skill.ClawsOfThunder, monster.UnitID, 4, opts...))
		}

		if !s.MobAlive(monster.UnitID, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Blades of Ice
		if !d.PlayerUnit.States.HasState(state.Bladesofice) {
			steps = append(steps, step.SecondaryAttack(skill.BladesOfIce, monster.UnitID, 4, opts...))
		}

		// Fists of Fire
		//if !d.PlayerUnit.States.HasState(state.Fistsoffire) {
		//	steps = append(steps, step.SecondaryAttack(skill.FistsOfFire, id, 4, opts...))
		//}

		if !s.MobAlive(monster.UnitID, d) { // Check if the mob is still alive
			return []step.Step{}
		}

		// Finish it off with primary attack
		steps = append(steps, step.PrimaryAttack(monster.UnitID, 2, opts...))

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s MosiacSin) KillCouncil() action.Action {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster

		for _, mobs := range d.Monsters.Enemies() {
			if mobs.Name == npc.CouncilMember || mobs.Name == npc.CouncilMember2 || mobs.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, mobs)
			}
		}

		for _, mobs := range councilMembers {
			return mobs.UnitID, true
		}

		return 0, false
	}, nil, step.Distance(1, 3))
}

func (s MosiacSin) KillBaal() action.Action {
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

func (s MosiacSin) KillIzual() action.Action {
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

func (s MosiacSin) KillDiablo() action.Action {
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

func (s MosiacSin) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s MosiacSin) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s MosiacSin) KillPindle(skipOnImmunities []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s MosiacSin) KillDuriel() action.Action {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s MosiacSin) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s MosiacSin) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s MosiacSin) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}
