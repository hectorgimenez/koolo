package character

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const (
	LightningMinDistance          = 10
	LightningMaxDistance          = 20
	LightningStaticMinDistance    = 1
	LightningStaticMaxDistance    = 3
	LightningMaxAttacksLoop       = 40
	LightningStaticFieldThreshold = 67 // Cast Static Field if monster HP is above this percentage
)

type LightningSorceress struct {
	BaseCharacter
}

func (s LightningSorceress) CheckKeyBindings() []skill.ID {
	requiredKeybindings := []skill.ID{skill.ChainLightning, skill.Teleport, skill.TomeOfTownPortal, skill.StaticField, skill.ShiverArmor}
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

func (s LightningSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()
	completedAttackLoops := 0
	staticFieldCast := false
	ldOpts := step.Distance(LightningMinDistance, LightningMaxDistance)
	lightningOpts := []step.AttackOption{
		step.RangedDistance(LightningMinDistance, LightningMaxDistance),
	}

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
				step.RangedDistance(LightningStaticMinDistance, LightningStaticMaxDistance),
			}

			if err := step.SecondaryAttack(skill.StaticField, monster.UnitID, 1, staticOpts...); err == nil {
				staticFieldCast = true
				continue
			}
		}

		if monster.Name == npc.Andariel ||
			monster.Name == npc.Duriel ||
			monster.Name == npc.Mephisto ||
			monster.Name == npc.Diablo ||
			monster.Name == npc.BaalCrab ||
			monster.Name == npc.Izual {
			if err := step.PrimaryAttack(monster.UnitID, 1, true, ldOpts); err == nil {
				completedAttackLoops++
			}
		} else {
			if err := step.SecondaryAttack(skill.ChainLightning, monster.UnitID, 1, lightningOpts...); err == nil {
				completedAttackLoops++
			}
		}

		if completedAttackLoops >= LightningMaxAttacksLoop {
			completedAttackLoops = 0
			staticFieldCast = false
		}
	}
}

func (s LightningSorceress) shouldCastStaticField(monster data.Monster) bool {
	// Only cast Static Field if monster HP is above threshold
	maxLife := float64(monster.Stats[stat.MaxLife])
	if maxLife == 0 {
		return false
	}

	hpPercentage := (float64(monster.Stats[stat.Life]) / maxLife) * 100
	return hpPercentage > LightningStaticFieldThreshold
}

func (s LightningSorceress) killBossWithStatic(bossID npc.ID, monsterType data.MonsterType) error {
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
				step.Distance(LightningStaticMinDistance, LightningStaticMaxDistance),
			}
			err := step.SecondaryAttack(skill.StaticField, boss.UnitID, 1, staticOpts...)
			if err != nil {
				s.Logger.Warn("Failed to cast Static Field", slog.String("error", err.Error()))
			}
			continue
		}

		// Switch to Lightning once boss HP is low enough
		return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			return boss.UnitID, true
		}, nil)
	}
}

func (s LightningSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s LightningSorceress) BuffSkills() []skill.ID {
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

func (s LightningSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s LightningSorceress) KillAndariel() error {
	return s.killBossWithStatic(npc.Andariel, data.MonsterTypeUnique)
}

func (s LightningSorceress) KillDuriel() error {
	return s.killBossWithStatic(npc.Duriel, data.MonsterTypeUnique)
}

func (s LightningSorceress) KillMephisto() error {
	return s.killBossWithStatic(npc.Mephisto, data.MonsterTypeUnique)
}

func (s LightningSorceress) KillDiablo() error {
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

func (s LightningSorceress) KillBaal() error {
	return s.killBossWithStatic(npc.BaalCrab, data.MonsterTypeUnique)
}

func (s LightningSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s LightningSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s LightningSorceress) KillIzual() error {
	return s.killBossWithStatic(npc.Izual, data.MonsterTypeUnique)
}

func (s LightningSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}

func (s LightningSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s LightningSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}
