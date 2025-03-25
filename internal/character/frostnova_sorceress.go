package character

import (
	"log/slog"
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

const (
	FrostNovaMinDistance    = 6
	FrostNovaMaxDistance    = 9
	FNStaticMinDistance     = 13
	FNStaticMaxDistance     = 22
	FrostNovaMaxAttacksLoop = 10
	FNStaticFieldThreshold  = 67 // Cast Static Field if monster HP is above this percentage
	FrostNovaOrbMinDistance = 6
	FrostNovaOrbMaxDistance = 15 // Left skill
)

type FrostNovaSorceress struct {
	BaseCharacter
}

func (s FrostNovaSorceress) CheckKeyBindings() []skill.ID {
	requiredKeybindings := []skill.ID{skill.FrostNova, skill.FrozenOrb, skill.Teleport, skill.TomeOfTownPortal, skill.StaticField}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requiredKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	// Check for one of the armor skills
	armorSkills := []skill.ID{skill.FrozenArmor, skill.ShiverArmor, skill.ChillingArmor}
	hasArmor := false
	for _, armor := range armorSkills {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(armor); found {
			hasArmor = true
			break
		}
	}
	if !hasArmor {
		missingKeybindings = append(missingKeybindings, skill.FrozenArmor)
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s FrostNovaSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()
	completedAttackLoops := 0
	staticFieldCast := false
	OrbOpts := step.Distance(FrostNovaOrbMinDistance, FrostNovaOrbMaxDistance)
	novaCastsSinceOrb := 0

	for {
		ctx.PauseIfNotPriority()

		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found || monster.Stats[stat.Life] <= 0 {
			return nil
		}

		// Cast Static Field first if needed
		if !staticFieldCast && s.shouldCastStaticField(monster) {
			staticOpts := []step.AttackOption{
				step.RangedDistance(FNStaticMinDistance, FNStaticMaxDistance),
			}

			if err := step.SecondaryAttack(skill.StaticField, monster.UnitID, 1, staticOpts...); err == nil {
				staticFieldCast = true
				continue
			}
		}

		if novaCastsSinceOrb >= 3 && !s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			if err := step.PrimaryAttack(id, 1, true, OrbOpts); err == nil {
				novaCastsSinceOrb = 0
				continue
			}
		}

		novaOpts := []step.AttackOption{
			step.RangedDistance(FrostNovaMinDistance, FrostNovaMaxDistance),
		}

		if s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			step.PrimaryAttack(id, 2, true, OrbOpts)
		}

		if err := step.SecondaryAttack(skill.FrostNova, monster.UnitID, 1, novaOpts...); err == nil {
			completedAttackLoops++
			novaCastsSinceOrb++
		}

		if completedAttackLoops >= FrostNovaMaxAttacksLoop {
			completedAttackLoops = 0
			staticFieldCast = false
		}
	}
}
func (s FrostNovaSorceress) shouldCastStaticField(monster data.Monster) bool {
	// Only cast Static Field if monster HP is above threshold
	maxLife := float64(monster.Stats[stat.MaxLife])
	if maxLife == 0 {
		return false
	}

	hpPercentage := (float64(monster.Stats[stat.Life]) / maxLife) * 100
	return hpPercentage > FNStaticFieldThreshold
}

func (s FrostNovaSorceress) killBossWithStatic(bossID npc.ID, monsterType data.MonsterType) error {
	ctx := context.Get()

	for {
		ctx.PauseIfNotPriority()

		boss, found := s.Data.Monsters.FindOne(bossID, monsterType)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil
		}

		bossHPPercent := (float64(boss.Stats[stat.Life]) / float64(boss.Stats[stat.MaxLife])) * 100
		thresholdFloat := float64(ctx.CharacterCfg.Character.NovaSorceress.BossStaticThreshold)

		// Cast Static Field until boss HP is below threshold
		if bossHPPercent > thresholdFloat {
			staticOpts := []step.AttackOption{
				step.Distance(FNStaticMinDistance, FNStaticMaxDistance),
			}
			err := step.SecondaryAttack(skill.StaticField, boss.UnitID, 1, staticOpts...)
			if err != nil {
				s.Logger.Warn("Failed to cast Static Field", slog.String("error", err.Error()))
			}
			continue
		}

		// Switch to Nova once boss HP is low enough
		return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			return boss.UnitID, true
		}, nil)
	}
}

func (s FrostNovaSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s FrostNovaSorceress) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.ThunderStorm); found {
		skillsList = append(skillsList, skill.ThunderStorm)
	}

	// Add one of the armor skills
	for _, armor := range []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor} {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			break
		}
	}

	return skillsList
}

func (s FrostNovaSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s FrostNovaSorceress) KillAndariel() error {
	return s.killBossWithStatic(npc.Andariel, data.MonsterTypeUnique)
}

func (s FrostNovaSorceress) KillDuriel() error {
	return s.killBossWithStatic(npc.Duriel, data.MonsterTypeUnique)
}

func (s FrostNovaSorceress) KillMephisto() error {
	return s.killBossWithStatic(npc.Mephisto, data.MonsterTypeUnique)
}

func (s FrostNovaSorceress) KillDiablo() error {
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

		return s.killBossWithStatic(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s FrostNovaSorceress) KillBaal() error {
	return s.killBossWithStatic(npc.BaalCrab, data.MonsterTypeUnique)
}

func (s FrostNovaSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s FrostNovaSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s FrostNovaSorceress) KillIzual() error {
	return s.killBossWithStatic(npc.Izual, data.MonsterTypeUnique)
}

func (s FrostNovaSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}

func (s FrostNovaSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s FrostNovaSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}
