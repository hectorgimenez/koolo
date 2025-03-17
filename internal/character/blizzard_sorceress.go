package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type BlizzardSorceress struct {
	BaseCharacter
}

type BlizzardAttackConfig struct {
	maxAttacksLoop    int
	attackMinDistance int
	attackMaxDistance int
}

func (s BlizzardSorceress) attackConfig() BlizzardAttackConfig {
	// Avoiding constants that pollute the other characters in the package
	return BlizzardAttackConfig{
		maxAttacksLoop:    40,
		attackMinDistance: 8,
		attackMaxDistance: 16,
	}
}

func (s BlizzardSorceress) CheckKeyBindings() []skill.ID {
	requireKeybindings := []skill.ID{skill.Blizzard, skill.Teleport, skill.TomeOfTownPortal, skill.ShiverArmor, skill.StaticField}

	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.ShiverArmor:
				_, found1 := s.Data.KeyBindings.KeyBindingForSkill(skill.FrozenArmor)
				_, found2 := s.Data.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
				if !found1 && !found2 {
					missingKeybindings = append(missingKeybindings, skill.ShiverArmor)
				}
			default:
				missingKeybindings = append(missingKeybindings, cskill)
			}
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s BlizzardSorceress) killMonsterWithStatic(bossID npc.ID, monsterType data.MonsterType) error {
	for {
		if !s.MonsterAliveByType(bossID, monsterType) {
			return nil
		}

		boss, _ := s.Data.Monsters.FindOne(bossID, monsterType)

		bossHPPercent := (float64(boss.Stats[stat.Life]) / float64(boss.Stats[stat.MaxLife])) * 100

		// Pull target threshold from config based on difficulty
		var targetThreshold int
		switch s.Data.CharacterCfg.Game.Difficulty {
		case difficulty.Normal:
			targetThreshold = 10
		case difficulty.Nightmare:
			targetThreshold = 40
		default:
			targetThreshold = 60
		}
		thresholdFloat := float64(targetThreshold)

		// Cast Static Field until boss HP is below threshold
		if bossHPPercent > thresholdFloat {
			staticOpts := []step.AttackOption{
				step.Distance(s.Data.CharacterCfg.Character.Sorceress.StaticFieldMinDist, s.Data.CharacterCfg.Character.Sorceress.StaticFieldMaxDist),
			}
			err := step.SecondaryAttack(skill.StaticField, boss.UnitID, 1, staticOpts...)
			if err != nil {
				s.Logger.Warn("Failed to cast Static Field", slog.String("error", err.Error()))
			}
			continue
		}

		return s.killMonsterByName(bossID, monsterType, nil)
	}
}

func (s BlizzardSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()
	completedAttackLoops := 0
	previousUnitID := 0
	previousSelfBlizzard := time.Time{}

	attackOpts := step.StationaryDistance(
		s.attackConfig().attackMinDistance,
		s.attackConfig().attackMaxDistance,
	)

	for {
		actionsTakenThisLoop := []string{}

		// Pause if not priority
		ctx.PauseIfNotPriority()

		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= s.attackConfig().maxAttacksLoop {
			s.Logger.Error("Exceeded MaxAttacksLoop", slog.String("completedAttackLoops", fmt.Sprintf("%v", completedAttackLoops)))
			return nil
		}

		for s.Data.PlayerUnit.States.HasState(state.Cooldown) && s.MonsterAliveById(id) {
			step.PrimaryAttack(id, 1, true, attackOpts)
			actionsTakenThisLoop = append(actionsTakenThisLoop, "PrimaryAttack")
			// Wait for the cast to complete before doing anything else
			time.Sleep(s.Data.PlayerCastDuration() - (120 * time.Millisecond)) // stolen from internal/action/step/attack.go
		}

		selfBlizzardThisLoop := false
		// Cast a Blizzard on very close mobs, in order to clear possible trash close to the player, every two attack rotations
		if time.Since(previousSelfBlizzard) > time.Second*4 {
			for _, m := range s.Data.Monsters.Enemies() {
				if dist := s.PathFinder.DistanceFromMe(m.Position); dist < 4 {
					previousSelfBlizzard = time.Now()
					selfBlizzardThisLoop = true
					step.SecondaryAttack(skill.Blizzard, m.UnitID, 1, attackOpts)
					actionsTakenThisLoop = append(actionsTakenThisLoop, "selfBlizz")
				}
			}
		}
		if !selfBlizzardThisLoop && s.MonsterAliveById(id) {
			step.SecondaryAttack(skill.Blizzard, id, 1, attackOpts)
			actionsTakenThisLoop = append(actionsTakenThisLoop, "offensiveBlizz")
		}

		if !s.MonsterAliveById(id) {
			return nil
		}

		completedAttackLoops++
		// s.Logger.Debug("Actions taken:", slog.String("actionsTakenThisLoop", fmt.Sprintf("%v", actionsTakenThisLoop)))

		previousUnitID = int(id)
	}
}

func (s BlizzardSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	// while the monster is alive, keep attacking it
	for {
		if m, found := s.Data.Monsters.FindOne(id, monsterType); found {
			if m.Stats[stat.Life] <= 0 {
				break
			}

			// Check if monster is immune to any of the skipOnImmunities
			for _, immunity := range skipOnImmunities {
				if m.IsImmune(immunity) {
					return nil
				}
			}

			s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				if m, found := d.Monsters.FindOne(id, monsterType); found {
					return m.UnitID, true
				}

				return 0, false
			}, skipOnImmunities)
		} else {
			break
		}
	}
	return nil
}

func (s BlizzardSorceress) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			return skillsList
		}
	}

	return skillsList
}

func (s BlizzardSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s BlizzardSorceress) KillCountess() error {
	return s.killMonsterWithStatic(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s BlizzardSorceress) KillAndariel() error {
	return s.killMonsterWithStatic(npc.Andariel, data.MonsterTypeUnique)
}

func (s BlizzardSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s BlizzardSorceress) KillDuriel() error {
	return s.killMonsterWithStatic(npc.Duriel, data.MonsterTypeUnique)
}

func (s BlizzardSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var coldImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.ColdImmune) {
					coldImmunes = append(coldImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}
		}

		councilMembers = append(councilMembers, coldImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s BlizzardSorceress) KillMephisto() error {
	return s.killMonsterWithStatic(npc.Mephisto, data.MonsterTypeUnique)
}

func (s BlizzardSorceress) KillIzual() error {
	return s.killMonsterWithStatic(npc.Izual, data.MonsterTypeUnique)
}

func (s BlizzardSorceress) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		if !s.MonsterAliveByType(npc.Diablo, data.MonsterTypeUnique) {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			time.Sleep(200)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")

		return s.killMonsterWithStatic(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s BlizzardSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s BlizzardSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (s BlizzardSorceress) KillBaal() error {
	return s.killMonsterWithStatic(npc.BaalCrab, data.MonsterTypeUnique)
}
