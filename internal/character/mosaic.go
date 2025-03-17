package character

import (
	"fmt"
	"log/slog"
	"slices"
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

func (s MosaicSin) chargedSkillStateMap() map[skill.ID]state.State {
	// Mappings of skills to the charge state
	return map[skill.ID]state.State{
		skill.TigerStrike:    state.State(stat.ProgressiveDamage),
		skill.CobraStrike:    state.State(stat.ProgressiveSteal),
		skill.PhoenixStrike:  state.State(stat.ProgressiveOther),
		skill.ClawsOfThunder: state.State(stat.ProgressiveLightning),
		skill.BladesOfIce:    state.State(stat.ProgressiveCold),
		skill.FistsOfFire:    state.State(stat.ProgressiveFire),
	}
}

func (s MosaicSin) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.PhoenixStrike, skill.TomeOfTownPortal}

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

func (s MosaicSin) hasKeyBindingForSkill(skill skill.ID) bool {
	_, found := s.Data.KeyBindings.KeyBindingForSkill(skill)
	return found
}

func (s MosaicSin) buildChargesForSkill(monsterId data.UnitID, skillToCharge skill.ID, desiredCount int, ctx context.Status) int {
	// Call this if we're enabled for the skill
	charges, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.ID(s.chargedSkillStateMap()[skillToCharge]), 0)
	attacks := 0

	if !s.MobAlive(monsterId, *s.Data) {
		return -1
	}

	if s.hasKeyBindingForSkill(skillToCharge) {
		if !found || (found && charges.Value < desiredCount) {
			attackCount := desiredCount - charges.Value
			attacks += attackCount
			step.SecondaryAttack(skillToCharge, monsterId, attackCount)
		}
	}

	return attacks
}

func (s MosaicSin) MobHasAnyState(mob data.UnitID, statesToFind []state.State) bool {
	monster, found := s.Data.Monsters.FindByID(mob)
	if found {
		for _, stateToFind := range statesToFind {
			if slices.Contains(monster.States, stateToFind) {
				return true
			}
		}
	}
	return false
}

func (s MosaicSin) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()
	ctx.RefreshGameData()
	lastRefresh := time.Now()

	// TODO: move to config
	attacksBeforeKick := 4
	useCloakOfShadows := true

	// Initial cloak of shadows cast for survivability
	if id, found := monsterSelector(*s.Data); found {
		if useCloakOfShadows && s.hasKeyBindingForSkill(skill.CloakOfShadows) &&
			!s.MobHasAnyState(id, []state.State{state.Lifetap, state.CloakOfShadows}) {
			// How do I determine if cloak of shadows is on cooldown?
			step.SecondaryAttack(skill.CloakOfShadows, id, 1)
		}
	}

	for {
		// Limit refresh rate to 10 times per second to avoid excessive CPU usage
		if time.Since(lastRefresh) > time.Millisecond*100 {
			ctx.RefreshGameData()
			lastRefresh = time.Now()
		}

		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		// Initial move to monster if we're too far
		if ctx.PathFinder.DistanceFromMe(monster.Position) > 3 {
			if s.hasKeyBindingForSkill(skill.DragonFlight) {
				step.SecondaryAttack(skill.DragonFlight, id, 1)
			} else if err := step.MoveTo(monster.Position); err != nil {
				s.Logger.Debug("Failed to move to monster position", slog.String("error", err.Error()))
				continue
			}
		}

		if !s.MobAlive(id, *s.Data) {
			return nil
		}

		// Cobra 3
		// Phoenix 2
		// Lightning 3
		// Tiger 3
		// Ice 3
		// Fire 3  // This one breaks life leech, dangerous?

		totalChargeAttacks := 0
		if ctx.CharacterCfg.Character.MosaicSin.UseCobraStrike && totalChargeAttacks < attacksBeforeKick {
			skillChargeCount := s.buildChargesForSkill(id, skill.CobraStrike, 3, *ctx)
			if skillChargeCount == -1 {
				return nil
			}
			totalChargeAttacks += skillChargeCount
		}

		// Always use phoenix
		if totalChargeAttacks < attacksBeforeKick {
			skillChargeCount := s.buildChargesForSkill(id, skill.PhoenixStrike, 2, *ctx)
			if skillChargeCount == -1 {
				return nil
			}
			totalChargeAttacks += skillChargeCount
		}

		if ctx.CharacterCfg.Character.MosaicSin.UseClawsOfThunder && totalChargeAttacks < attacksBeforeKick {
			skillChargeCount := s.buildChargesForSkill(id, skill.ClawsOfThunder, 3, *ctx)
			if skillChargeCount == -1 {
				return nil
			}
			totalChargeAttacks += skillChargeCount
		}

		if ctx.CharacterCfg.Character.MosaicSin.UseTigerStrike && totalChargeAttacks < attacksBeforeKick {
			skillChargeCount := s.buildChargesForSkill(id, skill.TigerStrike, 3, *ctx)
			if skillChargeCount == -1 {
				return nil
			}
			totalChargeAttacks += skillChargeCount
		}

		if ctx.CharacterCfg.Character.MosaicSin.UseBladesOfIce && totalChargeAttacks < attacksBeforeKick {
			skillChargeCount := s.buildChargesForSkill(id, skill.BladesOfIce, 3, *ctx)
			if skillChargeCount == -1 {
				return nil
			}
			totalChargeAttacks += skillChargeCount
		}

		// Note: you probably never want to use this. As fists of fire levels up, it converts more and
		// more of your physical damage to fire. You need physical damage for life leech!
		if ctx.CharacterCfg.Character.MosaicSin.UseFistsOfFire && totalChargeAttacks < attacksBeforeKick {
			skillChargeCount := s.buildChargesForSkill(id, skill.FistsOfFire, 3, *ctx)
			if skillChargeCount == -1 {
				return nil
			}
			totalChargeAttacks += skillChargeCount
		}

		opts := step.Distance(1, 2)
		// Finish it off with primary attack
		step.PrimaryAttack(id, 1, false, opts)
	}
}

func (s MosaicSin) MobAlive(mob data.UnitID, d game.Data) bool {
	monster, found := s.Data.Monsters.FindByID(mob)
	return found && monster.Stats[stat.Life] > 0
}

func (s MosaicSin) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)

	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.Fade); found {
		skillsList = append(skillsList, skill.Fade)
	} else {
		// If we don't use fade but we use Burst of Speed
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.BurstOfSpeed); found {
			skillsList = append(skillsList, skill.BurstOfSpeed)
		}
	}

	return skillsList
}

func (s MosaicSin) PreCTABuffSkills() []skill.ID {
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.ShadowMaster); found {
		return []skill.ID{skill.ShadowMaster}
	} else if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.ShadowWarrior); found {
		return []skill.ID{skill.ShadowWarrior}
	}
	return []skill.ID{}
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
	return s.killMonster(npc.Andariel, data.MonsterTypeUnique)
}

func (s MosaicSin) KillSummoner() error {
	return s.killMonster(npc.Summoner, data.MonsterTypeUnique)
}

func (s MosaicSin) KillDuriel() error {
	return s.killMonster(npc.Duriel, data.MonsterTypeUnique)
}

func (s MosaicSin) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := s.PathFinder.DistanceFromMe(councilMembers[i].Position)
			distanceJ := s.PathFinder.DistanceFromMe(councilMembers[j].Position)
			return distanceI < distanceJ
		})

		if len(councilMembers) > 0 {
			return councilMembers[0].UnitID, true
		}

		return 0, false
	}, nil)
}

func (s MosaicSin) KillMephisto() error {
	return s.killMonster(npc.Mephisto, data.MonsterTypeUnique)
}

func (s MosaicSin) KillIzual() error {
	return s.killMonster(npc.Izual, data.MonsterTypeUnique)
}

func (s MosaicSin) KillDiablo() error {
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
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")
		return s.killMonster(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s MosaicSin) KillPindle() error {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillNihlathak() error {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillBaal() error {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeUnique)
}
