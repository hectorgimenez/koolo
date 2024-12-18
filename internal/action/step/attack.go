package step

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/utils"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const attackCycleDuration = 120 * time.Millisecond

// Contains all configuration for an attack sequence
type attackSettings struct {
	primaryAttack    bool          // Whether this is a primary (left click) attack
	skill            skill.ID      // Skill ID for secondary attacks
	followEnemy      bool          // Whether to follow the enemy while attacking
	minDistance      int           // Minimum attack range
	maxDistance      int           // Maximum attack range
	aura             skill.ID      // Aura to maintain during attack
	target           data.UnitID   // Specific target's unit ID (0 for AOE)
	shouldStandStill bool          // Whether to stand still while attacking
	numOfAttacks     int           // Number of attacks to perform
	timeout          time.Duration // Timeout for the attack sequence
	isBurstCastSkill bool          // Whether this is a channeled/burst skill like Nova
}

// AttackOption defines a function type for configuring attack settings
type AttackOption func(step *attackSettings)

// Distance configures attack to follow enemy within specified range
func Distance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = true
		step.minDistance = minimum
		step.maxDistance = maximum
	}
}

// RangedDistance configures attack for ranged combat without following
func RangedDistance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = false // Don't follow enemies for ranged attacks
		step.minDistance = minimum
		step.maxDistance = maximum
	}
}

// StationaryDistance configures attack to remain stationary (like FoH)
func StationaryDistance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = false
		step.minDistance = minimum
		step.maxDistance = maximum
		step.shouldStandStill = true
	}
}

// EnsureAura ensures specified aura is active during attack
func EnsureAura(aura skill.ID) AttackOption {
	return func(step *attackSettings) {
		step.aura = aura
	}
}

// PrimaryAttack initiates a primary (left-click) attack sequence
func PrimaryAttack(target data.UnitID, numOfAttacks int, standStill bool, opts ...AttackOption) error {
	ctx := context.Get()

	// Special handling for Berserker characters
	if berserker, ok := ctx.Char.(interface{ PerformBerserkAttack(data.UnitID) }); ok {
		for i := 0; i < numOfAttacks; i++ {
			berserker.PerformBerserkAttack(target)
		}
		return nil
	}

	settings := attackSettings{
		target:           target,
		numOfAttacks:     numOfAttacks,
		shouldStandStill: standStill,
		primaryAttack:    true,
	}
	for _, o := range opts {
		o(&settings)
	}

	return attack(settings)
}

// SecondaryAttack initiates a secondary (right-click) attack sequence with a specific skill
func SecondaryAttack(skill skill.ID, target data.UnitID, numOfAttacks int, opts ...AttackOption) error {
	settings := attackSettings{
		target:           target,
		numOfAttacks:     numOfAttacks,
		skill:            skill,
		primaryAttack:    false,
		isBurstCastSkill: skill == 48, // nova can define any other burst skill here
	}
	for _, o := range opts {
		o(&settings)
	}

	if settings.isBurstCastSkill {
		settings.timeout = 30 * time.Second
		return burstAttack(settings)
	}

	return attack(settings)
}

// Helper function to validate if a monster should be targetable
func isValidEnemy(monster data.Monster, ctx *context.Status) bool {
	// Special case: Always allow Vizier seal boss even if off grid
	isVizier := monster.Type == data.MonsterTypeSuperUnique && monster.Name == npc.StormCaster
	if isVizier {
		return monster.Stats[stat.Life] > 0
	}

	// Skip monsters in invalid positions
	if !ctx.Data.AreaData.IsWalkable(monster.Position) {
		ctx.Logger.Debug("Skipping monster in unwalkable position",
			slog.Any("monster", monster.Name),
			slog.Any("position", monster.Position))
		return false
	}

	// Skip dead monsters
	if monster.Stats[stat.Life] <= 0 {
		return false
	}

	return true
}

// Cleanup function to ensure proper state on exit
func keyCleanup(ctx *context.Status) {
	ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
}

func attack(settings attackSettings) error {
	ctx := context.Get()
	ctx.SetLastStep("Attack")
	defer keyCleanup(ctx) // cleanup possible pressed keys/buttons

	numOfAttacksRemaining := settings.numOfAttacks

	lastRunAt := time.Time{}
	for {
		ctx.PauseIfNotPriority()

		if numOfAttacksRemaining <= 0 {
			return nil
		}

		monster, found := ctx.Data.Monsters.FindByID(settings.target)
		if !found || !isValidEnemy(monster, ctx) {
			return nil // Target is not valid, we don't have anything to attack
		}

		distance := ctx.PathFinder.DistanceFromMe(monster.Position)
		if !lastRunAt.IsZero() && !settings.followEnemy && distance > settings.maxDistance {
			return nil // Enemy is out of range and followEnemy is disabled, we cannot attack
		}

		// Be sure we stay in range of the enemy
		err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance)
		if err != nil {
			return fmt.Errorf("enemy is out of range and cannot be reached: %w", err)
		}

		// Handle aura activation
		if settings.aura != 0 && lastRunAt.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		// Attack timing check
		if time.Since(lastRunAt) <= ctx.Data.PlayerCastDuration()-attackCycleDuration {
			continue
		}

		performAttack(ctx, settings, monster.Position.X, monster.Position.Y)

		lastRunAt = time.Now()
		numOfAttacksRemaining--
	}
}

func burstAttack(settings attackSettings) error {
	ctx := context.Get()
	ctx.SetLastStep("BurstAttack")
	defer keyCleanup(ctx) // cleanup possible pressed keys/buttons

	monster, found := ctx.Data.Monsters.FindByID(settings.target)
	if !found || !isValidEnemy(monster, ctx) {
		return nil // Target is not valid, we don't have anything to attack
	}

	// Initially we try to move to the enemy, later we will check for closer enemies to keep attacking
	err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance)
	if err != nil {
		return fmt.Errorf("enemy is out of range and cannot be reached: %w", err)
	}

	startedAt := time.Time{}
	for {
		ctx.PauseIfNotPriority()

		if !startedAt.IsZero() && time.Since(startedAt) > settings.timeout {
			return nil // Timeout reached, finish attack sequence
		}

		target := data.Monster{}
		for _, monster = range ctx.Data.Monsters.Enemies() {
			distance := ctx.PathFinder.DistanceFromMe(monster.Position)
			if isValidEnemy(monster, ctx) && distance <= settings.maxDistance {
				target = monster
				break
			}
		}

		if target.UnitID == 0 {
			return nil // We have no valid targets in range, finish attack sequence
		}

		// If we don't have LoS we will need to interrupt and move :(
		if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, target.Position) {
			err = ensureEnemyIsInRange(target, settings.maxDistance, settings.minDistance)
			if err != nil {
				return fmt.Errorf("enemy is out of range and cannot be reached: %w", err)
			}
		}

		performAttack(ctx, settings, target.Position.X, target.Position.Y)
	}
}

func performAttack(ctx *context.Status, settings attackSettings, x, y int) {
	// Ensure we have the skill selected
	if settings.skill != 0 && ctx.Data.PlayerUnit.RightSkill != settings.skill {
		ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
		time.Sleep(time.Millisecond * 10)
	}

	if settings.shouldStandStill {
		ctx.HID.KeyDown(ctx.Data.KeyBindings.StandStill)
	}

	x, y = ctx.PathFinder.GameCoordsToScreenCords(x, y)
	if settings.primaryAttack {
		ctx.HID.Click(game.LeftButton, x, y)
	} else {
		ctx.HID.Click(game.RightButton, x, y)
	}

	if settings.shouldStandStill {
		ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
	}
}

func ensureEnemyIsInRange(monster data.Monster, maxDistance, minDistance int) error {
	ctx := context.Get()
	ctx.SetLastStep("ensureEnemyIsInRange")

	// TODO: Add an option for telestomp based on the char configuration
	currentPos := ctx.Data.PlayerUnit.Position
	distanceToMonster := ctx.PathFinder.DistanceFromMe(monster.Position)
	hasLoS := ctx.PathFinder.LineOfSight(currentPos, monster.Position)

	// We have line of sight, and we are inside the attack range, we can skip
	if hasLoS && distanceToMonster <= maxDistance && distanceToMonster >= minDistance {
		return nil
	}

	// Get path to monster
	path, _, found := ctx.PathFinder.GetPath(monster.Position)
	// We cannot reach the enemy, let's skip the attack sequence
	if !found {
		return errors.New("path could not be calculated")
	}

	// Any close-range combat (mosaic,barb...) should move directly to target
	if maxDistance <= 3 {
		return MoveTo(monster.Position)
	}

	// Look for suitable position along path
	for _, pos := range path {
		monsterDistance := utils.DistanceFromPoint(ctx.Data.AreaData.RelativePosition(monster.Position), pos)
		if monsterDistance > maxDistance || monsterDistance < minDistance {
			continue
		}

		dest := data.Position{
			X: pos.X + ctx.Data.AreaData.OffsetX,
			Y: pos.Y + ctx.Data.AreaData.OffsetY,
		}

		// Calculate how far we need to move to reach this position
		distanceToMove := ctx.PathFinder.DistanceFromMe(dest)

		// If we need to move less than 7 units, we need to overshoot
		if distanceToMove <= 7 {
			// Calculate vector from current pos to destination
			dx := float64(dest.X - currentPos.X)
			dy := float64(dest.Y - currentPos.Y)

			// Normalize and extend to 9 units (beyond the 7 unit minimum)
			length := math.Sqrt(dx*dx + dy*dy)
			if length == 0 {
				dx = 1
				length = 1
			}
			dx = dx / length * 9
			dy = dy / length * 9

			// Create new overshooting destination
			dest = data.Position{
				X: currentPos.X + int(dx),
				Y: currentPos.Y + int(dy),
			}
		}

		if ctx.PathFinder.LineOfSight(dest, monster.Position) {
			return MoveTo(dest)
		}
	}

	return nil
}
