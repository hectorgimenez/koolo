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
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type MosaicSin struct {
	BaseCharacter
}

func (s MosaicSin) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.TigerStrike, skill.CobraStrike, skill.PhoenixStrike, skill.ClawsOfThunder, skill.BladesOfIce, skill.TomeOfTownPortal}
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

func (s MosaicSin) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()
	ctx.RefreshGameData()
	lastRefresh := time.Now()

	for {
		// Limit refresh rate to 10 times per second to avoid excessive CPU usage
		if time.Since(lastRefresh) > time.Millisecond*100 {
			ctx.RefreshGameData()
			lastRefresh = time.Now()
		}

		// Get the charges for each skill we're using
		tigerCharges, foundTiger := ctx.Data.PlayerUnit.Stats.FindStat(stat.ProgressiveDamage, 0)
		cobraCharges, foundCobra := ctx.Data.PlayerUnit.Stats.FindStat(stat.ProgressiveSteal, 0)
		phoenixCharges, foundPhoenix := ctx.Data.PlayerUnit.Stats.FindStat(stat.ProgressiveOther, 0)
		clawsCharges, foundClaws := ctx.Data.PlayerUnit.Stats.FindStat(stat.ProgressiveLightning, 0)
		bladesCharges, foundBlades := ctx.Data.PlayerUnit.Stats.FindStat(stat.ProgressiveCold, 0)

		id, found := monsterSelector(*s.data)
		if !found {
			return nil
		}

		monster, found := s.data.Monsters.FindByID(id)
		if !found {
			s.logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		opts := step.Distance(1, 2)

		if !s.MobAlive(id, *s.data) { // Check if the mob is still alive
			return nil
		}

		// Number of charges display as stats
		/*
			318 = Tiger Strike
			319 = Cobra Strike
			320 = Phoenix Strike
			321 = Fists of Fire
			322 = Blades of Ice
			323 = Claws of Thunder
		*/

		// Tiger Strike - 3 charges
		if !s.data.PlayerUnit.States.HasState(state.Tigerstrike) || (foundTiger && tigerCharges.Value < 3) {
			step.SecondaryAttack(skill.TigerStrike, id, 1, opts)
			continue
		}

		if !s.MobAlive(id, *s.data) { // Check if the mob is still alive
			return nil
		}

		// Cobra Strike - 3 charges
		if !s.data.PlayerUnit.States.HasState(state.Cobrastrike) || (foundCobra && cobraCharges.Value < 3) {
			step.SecondaryAttack(skill.CobraStrike, id, 1, opts)
			continue
		}

		if !s.MobAlive(id, *s.data) { // Check if the mob is still alive
			return nil
		}

		// Phoenix Strike - 2 charges
		if !s.data.PlayerUnit.States.HasState(state.Phoenixstrike) || (foundPhoenix && phoenixCharges.Value < 2) {
			step.SecondaryAttack(skill.PhoenixStrike, id, 1, opts)
			continue
		}

		if !s.MobAlive(id, *s.data) { // Check if the mob is still alive
			return nil
		}

		// Claws of Thunder - 3 charges
		if !s.data.PlayerUnit.States.HasState(state.Clawsofthunder) || (foundClaws && clawsCharges.Value < 3) {
			step.SecondaryAttack(skill.ClawsOfThunder, id, 1, opts)
			continue
		}

		if !s.MobAlive(id, *s.data) { // Check if the mob is still alive
			return nil
		}

		// Blades of Ice - 3 charges
		if !s.data.PlayerUnit.States.HasState(state.Bladesofice) || (foundBlades && bladesCharges.Value < 3) {
			step.SecondaryAttack(skill.BladesOfIce, id, 1, opts)
			continue
		}

		if !s.MobAlive(id, *s.data) { // Check if the mob is still alive
			return nil
		}

		// Finish it off with primary attack
		step.PrimaryAttack(id, 1, false, opts)
	}
}

func (s MosaicSin) MobAlive(mob data.UnitID, d game.Data) bool {
	monster, found := s.data.Monsters.FindByID(mob)
	return found && monster.Stats[stat.Life] > 0
}

func (s MosaicSin) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)

	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.Fade); found {
		skillsList = append(skillsList, skill.Fade)
	} else {

		// If we don't use fade but we use Burst of Speed
		if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.BurstOfSpeed); found {
			skillsList = append(skillsList, skill.BurstOfSpeed)
		}
	}

	return skillsList
}

func (s MosaicSin) PreCTABuffSkills() []skill.ID {

	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.ShadowMaster); found {
		return []skill.ID{skill.ShadowMaster}
	} else if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.ShadowWarrior); found {
		return []skill.ID{skill.ShadowWarrior}
	} else {
		return []skill.ID{}
	}

}

func (s MosaicSin) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s MosaicSin) KillCountess() error {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillAndariel() error {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s MosaicSin) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s MosaicSin) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s MosaicSin) KillCouncil() error {
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

func (s MosaicSin) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s MosaicSin) KillIzual() error {
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)
	s.killMonster(npc.Izual, data.MonsterTypeNone)

	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s MosaicSin) KillDiablo() error {
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

func (s MosaicSin) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillBaal() error {
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
	s.killMonster(npc.BaalCrab, data.MonsterTypeNone)

	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}
