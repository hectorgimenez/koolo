package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

const (
	NovaSorceressMaxAttacksLoop = 10
	NovaSorceressMinDistance    = 6
	NovaSorceressMaxDistance    = 10
	StaticFieldMaxDistance      = 13
)

type NovaSorceress struct {
	BaseCharacter
}

func (s NovaSorceress) CheckKeyBindings() []skill.ID {
	requiredKeybindings := []skill.ID{skill.Nova, skill.Teleport, skill.TomeOfTownPortal, skill.StaticField}
	missingKeybindings := []skill.ID{}

	for _, cskill := range requiredKeybindings {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(cskill); !found {
			missingKeybindings = append(missingKeybindings, cskill)
		}
	}

	// Check for one of the armor skills
	armorSkills := []skill.ID{skill.FrozenArmor, skill.ShiverArmor, skill.ChillingArmor}
	hasArmor := false
	for _, armor := range armorSkills {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(armor); found {
			hasArmor = true
			break
		}
	}
	if !hasArmor {
		missingKeybindings = append(missingKeybindings, skill.FrozenArmor)
	}

	if len(missingKeybindings) > 0 {
		s.logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s NovaSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	completedAttackLoops := 0
	previousUnitID := 0

	for {
		id, found := monsterSelector(*s.data)
		if !found {
			return nil
		}
		ctx := context.Get()
		ctx.PauseIfNotPriority()

		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= NovaSorceressMaxAttacksLoop {
			return nil
		}

		monster, found := s.data.Monsters.FindByID(id)
		if !found {
			s.logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		distance := s.pf.DistanceFromMe(monster.Position)
		if distance > NovaSorceressMaxDistance {
			safePosition := s.findSafePosition(monster.Position, NovaSorceressMinDistance, NovaSorceressMaxDistance)
			if safePosition != s.data.PlayerUnit.Position {
				err := step.MoveTo(safePosition)
				if err != nil {
					s.logger.Warn("Failed to move closer to monster", slog.String("error", err.Error()))
				}
			}
			continue
		}

		opts := step.Distance(NovaSorceressMinDistance, NovaSorceressMaxDistance)

		if s.shouldCastStatic() {
			step.SecondaryAttack(skill.StaticField, id, 1, opts)
		}

		step.SecondaryAttack(skill.Nova, id, 3, opts)

		completedAttackLoops++
		previousUnitID = int(id)

		// Add a small delay between attacks
		time.Sleep(50 * time.Millisecond)
	}
}
func (s NovaSorceress) killBossWithStatic(bossID npc.ID, monsterType data.MonsterType) error {
	ctx := context.Get()
	ctx.PauseIfNotPriority()
	for {
		boss, found := s.data.Monsters.FindOne(bossID, monsterType)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil
		}

		bossHPPercent := (float64(boss.Stats[stat.Life]) / float64(boss.Stats[stat.MaxLife])) * 100

		// Move closer if too far for Static Field
		distance := s.pf.DistanceFromMe(boss.Position)
		if distance > StaticFieldMaxDistance {
			err := step.MoveTo(boss.Position)
			if err != nil {
				s.logger.Warn("Failed to move closer to boss", slog.String("error", err.Error()))
			}
			utils.Sleep(100) // Short delay after moving
		}
		// Convert BossStaticThreshold to float64 before comparison
		thresholdFloat := float64(ctx.CharacterCfg.Character.NovaSorceress.BossStaticThreshold)
		if bossHPPercent > thresholdFloat {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.StaticField))
			utils.Sleep(80)
			x, y := ctx.PathFinder.GameCoordsToScreenCords(boss.Position.X, boss.Position.Y)
			ctx.HID.Click(game.RightButton, x, y)
			utils.Sleep(150)
		} else {
			// Switch to Nova and attack
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.Nova))
			utils.Sleep(80) // Short delay to ensure skill switch
			return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				return boss.UnitID, true
			}, nil)
		}
	}
}
func (s NovaSorceress) findSafePosition(monsterPos data.Position, minDistance, maxDistance int) data.Position {
	playerPos := s.data.PlayerUnit.Position
	currentDistance := s.pf.DistanceFromMe(monsterPos)

	if currentDistance >= minDistance && currentDistance <= maxDistance {
		return playerPos // Already at a safe distance
	}

	// Check positions at increasing distances from the monster
	for distance := minDistance; distance <= maxDistance; distance++ {
		positions := s.getPositionsAtDistance(monsterPos, distance)
		for _, pos := range positions {
			if s.data.AreaData.Grid.IsWalkable(pos) {
				return pos
			}
		}
	}

	return playerPos // If no safe position found, don't move
}

func (s NovaSorceress) getPositionsAtDistance(center data.Position, distance int) []data.Position {
	positions := make([]data.Position, 0, 8*distance)
	for x := -distance; x <= distance; x++ {
		positions = append(positions, data.Position{X: center.X + x, Y: center.Y + distance})
		positions = append(positions, data.Position{X: center.X + x, Y: center.Y - distance})
	}
	for y := -distance + 1; y < distance; y++ {
		positions = append(positions, data.Position{X: center.X + distance, Y: center.Y + y})
		positions = append(positions, data.Position{X: center.X - distance, Y: center.Y + y})
	}
	return positions
}

func (s NovaSorceress) killMonster(npc npc.ID, t data.MonsterType) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return 0, false
		}

		return m.UnitID, true
	}, nil)
}

func (s NovaSorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, maxDistance int, _ bool, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s NovaSorceress) shouldCastStatic() bool {
	nearbyMobs := make([]data.Monster, 0)
	for _, monster := range s.data.Monsters.Enemies() {
		if s.pf.DistanceFromMe(monster.Position) <= NovaSorceressMaxDistance {
			nearbyMobs = append(nearbyMobs, monster)
		}
	}

	if len(nearbyMobs) == 0 {
		return false
	}

	mobsAbove60Percent := 0
	for _, mob := range nearbyMobs {
		lifePercentage := int((float64(mob.Stats[stat.Life]) / float64(mob.Stats[stat.MaxLife])) * 100)
		if lifePercentage > 60 {
			mobsAbove60Percent++
		}
	}

	return mobsAbove60Percent > len(nearbyMobs)/2
}

func (s NovaSorceress) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}
	if _, found := s.data.KeyBindings.KeyBindingForSkill(skill.ThunderStorm); found {
		skillsList = append(skillsList, skill.ThunderStorm)
	}

	// Add one of the armor skills
	for _, armor := range []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor} {
		if _, found := s.data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			break
		}
	}

	return skillsList
}

func (s NovaSorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s NovaSorceress) KillAndariel() error {
	for {
		boss, found := s.data.Monsters.FindOne(npc.Andariel, data.MonsterTypeUnique)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil // Andariel is dead or not found
		}

		err := s.killBossWithStatic(npc.Andariel, data.MonsterTypeUnique)
		if err != nil {
			return err
		}

		// Short delay before checking again
		time.Sleep(100 * time.Millisecond)
	}
}

func (s NovaSorceress) KillDuriel() error {
	return s.killBossWithStatic(npc.Duriel, data.MonsterTypeUnique)
}

func (s NovaSorceress) KillMephisto() error {
	for {
		boss, found := s.data.Monsters.FindOne(npc.Mephisto, data.MonsterTypeUnique)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil // Mephisto is dead or not found
		}

		err := s.killBossWithStatic(npc.Mephisto, data.MonsterTypeUnique)
		if err != nil {
			return err
		}

		// Short delay before checking again
		time.Sleep(100 * time.Millisecond)
	}
}

func (s NovaSorceress) KillDiablo() error {
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			if diabloFound {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return s.killBossWithStatic(npc.Diablo, data.MonsterTypeUnique)
	}
}

func (s NovaSorceress) KillBaal() error {
	return s.killBossWithStatic(npc.BaalCrab, data.MonsterTypeUnique)
}

func (s NovaSorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, NovaSorceressMaxDistance, false, nil)
}

func (s NovaSorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, NovaSorceressMaxDistance, false, nil)
}

func (s NovaSorceress) KillIzual() error {
	return s.killBossWithStatic(npc.Izual, data.MonsterTypeUnique)
}

func (s NovaSorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				return m.UnitID, true
			}
		}
		return 0, false
	}, nil)
}

func (s NovaSorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, sorceressMaxDistance, false, s.cfg.Game.Pindleskin.SkipOnImmunities)
}

func (s NovaSorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, NovaSorceressMaxDistance, false, nil)
}
