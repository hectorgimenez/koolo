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

type ChargeSkillConfigEntry struct {
	skill            skill.ID
	useSkill         bool
	chargeState      state.State
	chargesPerAttack int
	desiredCharges   int
}

func (s MosaicSin) getChargeSkillConfig(ctx context.Status) map[skill.ID]ChargeSkillConfigEntry {
	return map[skill.ID]ChargeSkillConfigEntry{
		skill.TigerStrike: {
			skill.TigerStrike,
			ctx.CharacterCfg.Character.MosaicSin.UseTigerStrike,
			state.State(stat.ProgressiveDamage),
			1,
			3,
		},
		skill.CobraStrike: {
			skill.CobraStrike,
			ctx.CharacterCfg.Character.MosaicSin.UseCobraStrike,
			state.State(stat.ProgressiveSteal),
			1,
			3,
		},
		skill.PhoenixStrike: {
			skill.PhoenixStrike,
			true, // Always enabled
			state.State(stat.ProgressiveOther),
			1,
			2, // Two charges, we want lightning
		},
		skill.ClawsOfThunder: {
			skill.ClawsOfThunder,
			ctx.CharacterCfg.Character.MosaicSin.UseClawsOfThunder,
			state.State(stat.ProgressiveLightning),
			2,
			3,
		},
		skill.BladesOfIce: {
			skill.BladesOfIce,
			ctx.CharacterCfg.Character.MosaicSin.UseBladesOfIce,
			state.State(stat.ProgressiveCold),
			2,
			3,
		},
		// Note: Be careful with this. As fists of fire levels up, it converts more and
		// more of your physical damage to fire. You need physical damage for life leech!
		skill.FistsOfFire: {
			skill.FistsOfFire,
			ctx.CharacterCfg.Character.MosaicSin.UseFistsOfFire,
			state.State(stat.ProgressiveFire),
			2,
			3,
		},
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

func (s MosaicSin) buildChargesForSkill(monsterId data.UnitID, skillConfig ChargeSkillConfigEntry, ctx context.Status) (int, bool) {
	// The configuration checks for whether this skill is enabled are handled before we call this function

	charges, found := ctx.Data.PlayerUnit.Stats.FindStat(stat.ID(skillConfig.chargeState), 0)
	attacks := 0

	if !s.MonsterAliveById(monsterId) {
		return -1, true
	}

	if s.hasKeyBindingForSkill(skillConfig.skill) {
		if !found || (found && charges.Value < skillConfig.desiredCharges) {
			neededCharges := skillConfig.desiredCharges - charges.Value
			// Some skills give up to 2 charges per attack
			plannedAttacks := (neededCharges + skillConfig.chargesPerAttack - 1) / skillConfig.chargesPerAttack
			attacks += plannedAttacks
			step.SecondaryAttack(skillConfig.skill, monsterId, plannedAttacks, step.MeleeDistance(2))
		}
	}

	return attacks, false
}

func (s MosaicSin) MobHasAnyState(mob data.UnitID, statesToFind []state.State) bool {
	// TODO: this maybe has a home in d2go?
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

func (s MosaicSin) AttackLoop(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	attackOrder []skill.ID,
	skillConfig map[skill.ID]ChargeSkillConfigEntry,
	ctx context.Status,
) error {
	ctx.Data.PlayerUnit = ctx.GameReader.GetData().Data.PlayerUnit
	lastRefresh := time.Now()

	// TODO: move to config?
	attacksBeforeKick := 4
	useCloakOfShadows := true

	// Initial cloak of shadows cast for survivability
	if id, found := monsterSelector(*s.Data); found {
		// How do I determine if cloak of shadows is on cooldown?
		if useCloakOfShadows && s.hasKeyBindingForSkill(skill.CloakOfShadows) &&
			!s.MobHasAnyState(id, []state.State{state.Lifetap, state.CloakOfShadows}) {
			step.SecondaryAttack(skill.CloakOfShadows, id, 1, step.StationaryDistance(1, 24))
		}
	}

	for {
		// Limit refresh rate to 10 times per second to avoid excessive CPU usage
		if time.Since(lastRefresh) > time.Millisecond*100 {
			ctx.Data.PlayerUnit = ctx.GameReader.GetData().Data.PlayerUnit
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

		if !s.MonsterAliveById(id) {
			return nil
		}

		totalAttacks := 0

		for _, chargeSkill := range attackOrder {
			if skillConfig[chargeSkill].useSkill && totalAttacks < attacksBeforeKick {
				attackCount, alreadyDead := s.buildChargesForSkill(
					id,
					skillConfig[chargeSkill],
					ctx,
				)

				if alreadyDead {
					return nil
				}

				totalAttacks += attackCount
			}
		}

		// Finish it off with primary attack
		step.PrimaryAttack(id, 1, false, step.MeleeDistance(4))
	}
}

func (s MosaicSin) KillBossSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	chargeSkills := []skill.ID{
		// skill.CobraStrike,
		skill.PhoenixStrike, // We only want to use phoenix strike for bosses, but you can uncomment and reorder these if you desire.
		// skill.ClawsOfThunder,
		// skill.TigerStrike,
		// skill.BladesOfIce,
		// skill.FistsOfFire,
	}

	ctx := context.Get()
	chargeSkillConfig := s.getChargeSkillConfig(*ctx)
	// Override phoenix strike behavior for bosses
	chargeSkillConfig[skill.PhoenixStrike] = ChargeSkillConfigEntry{
		skill.PhoenixStrike,
		true, // Always enabled
		state.State(stat.ProgressiveOther),
		1,
		1, // We only want one charge of Phoenix strike for bosses
	}

	return s.AttackLoop(
		monsterSelector,
		skipOnImmunities,
		chargeSkills,
		chargeSkillConfig,
		*ctx,
	)
}

func (s MosaicSin) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	// The order to charge our skills, if the skill is enabled. TODO: Order could be configurable
	chargeSkills := []skill.ID{
		skill.CobraStrike,
		skill.PhoenixStrike,
		skill.ClawsOfThunder,
		skill.TigerStrike,
		skill.BladesOfIce,
		skill.FistsOfFire,
	}

	ctx := context.Get()
	chargeSkillConfig := s.getChargeSkillConfig(*ctx)

	return s.AttackLoop(
		monsterSelector,
		skipOnImmunities,
		chargeSkills,
		chargeSkillConfig,
		*ctx,
	)
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

func (s MosaicSin) killBoss(npc npc.ID, t data.MonsterType) error {
	return s.KillBossSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}
		return m.UnitID, true
	}, nil)
}

func (s MosaicSin) KillCountess() error {
	return s.killBoss(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillAndariel() error {
	return s.killBoss(npc.Andariel, data.MonsterTypeUnique)
}

func (s MosaicSin) KillSummoner() error {
	return s.killBoss(npc.Summoner, data.MonsterTypeUnique)
}

func (s MosaicSin) KillDuriel() error {
	return s.killBoss(npc.Duriel, data.MonsterTypeUnique)
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
	return s.killBoss(npc.Mephisto, data.MonsterTypeUnique)
}

func (s MosaicSin) KillIzual() error {
	return s.killBoss(npc.Izual, data.MonsterTypeUnique)
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
		return s.killBoss(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s MosaicSin) KillPindle() error {
	return s.killBoss(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillNihlathak() error {
	return s.killBoss(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s MosaicSin) KillBaal() error {
	return s.killBoss(npc.BaalCrab, data.MonsterTypeUnique)
}
