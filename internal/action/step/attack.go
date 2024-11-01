package step

import (
	"errors"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data/npc"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const attackCycleDuration = 120 * time.Millisecond

// Contains all configuration for an attack sequence
type attackSettings struct {
	primaryAttack     bool        // Whether this is a primary (left click) attack
	skill             skill.ID    // Skill ID for secondary attacks
	followEnemy       bool        // Whether to follow the enemy while attacking
	minDistance       int         // Minimum attack range
	maxDistance       int         // Maximum attack range
	telestomp         bool        // Whether to use teleport stomping
	aura              skill.ID    // Aura to maintain during attack
	target            data.UnitID // Specific target's unit ID (0 for AOE)
	shouldStandStill  bool        // Whether to stand still while attacking
	numOfAttacks      int         // Number of attacks to perform
	isBurstCastSkill  bool        // Whether this is a channeled/burst skill like Nova
	isMeleeAttack     bool        // Whether this is a melee range attack
	currentlyBursting bool        // Whether currently performing a burst cast
	isAOE             bool        // Whether this is aoe attack
}

// AttackOption defines a function type for configuring attack settings
type AttackOption func(step *attackSettings)

// Distance configures attack to follow enemy within specified range
func Distance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = true
		step.minDistance = minimum
		step.maxDistance = maximum
		// Set isMeleeAttack if it's melee range
		step.isMeleeAttack = minimum <= 1 && maximum <= 3
	}
}

// RangedDistance configures attack for ranged combat without following
func RangedDistance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = false // Don't follow enemies for ranged attacks
		step.minDistance = minimum
		step.maxDistance = maximum
		step.isMeleeAttack = false
	}
}

// StationaryDistance configures attack to remain stationary (like FoH)
func StationaryDistance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = false
		step.minDistance = minimum
		step.maxDistance = maximum
		step.isMeleeAttack = false
		step.shouldStandStill = true
	}
}

// EnsureAura ensures specified aura is active during attack
func EnsureAura(aura skill.ID) AttackOption {
	return func(step *attackSettings) {
		step.aura = aura
	}
}

// Telestomp will attempt attacking with mercenary by Teleporting on target
func Telestomp() AttackOption {
	return func(step *attackSettings) {
		step.telestomp = true
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
		isAOE:            skill == 48, // nova, can define other AOE skills here
	}
	for _, o := range opts {
		o(&settings)
	}

	return attack(settings)
}

// Helper function to validate if a monster should be targetable
func isValidTarget(monster data.Monster, ctx *context.Status) bool {
	// Skip off-grid monsters
	if !ctx.Data.AreaData.IsInside(monster.Position) {
		ctx.Logger.Debug("Skipping off-grid monster",
			slog.Any("monster", monster.Name),
			slog.Any("position", monster.Position))
		return false
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

// attack performs the main attack sequence based on provided settings
func attack(settings attackSettings) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "Attack"

	numOfAttacksRemaining := settings.numOfAttacks
	aoe := settings.target == 0
	lastRun := time.Time{}
	initialPosition := ctx.Data.PlayerUnit.Position

	// Cleanup function to ensure proper state on exit
	cleanup := func() {
		ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
		ctx.HID.ReleaseMouseButton(game.RightButton)
	}
	defer cleanup()

	// Skip monsters that are off grid unless it's a special case
	isMonsterValid := func(monster data.Monster) bool {
		// Special case: Always allow Vizier seal boss even if off grid
		isVizier := monster.Type == data.MonsterTypeSuperUnique && monster.Name == npc.StormCaster
		if isVizier {
			return monster.Stats[stat.Life] > 0
		}

		return isValidTarget(monster, ctx)
	}

	// Check if we have any valid targets within range
	hasValidTargets := func() bool {
		if !aoe {
			monster, found := ctx.Data.Monsters.FindByID(settings.target)
			if !found {
				return false
			}
			return isMonsterValid(monster)
		}

		// For AoE skills, check all monsters in range
		for _, monster := range ctx.Data.Monsters.Enemies() {
			if !isMonsterValid(monster) {
				continue
			}
			distance := ctx.PathFinder.DistanceFromMe(monster.Position)
			if distance >= settings.minDistance && distance <= settings.maxDistance {
				return true
			}
		}
		return false
	}

	for {
		ctx.PauseIfNotPriority()

		if numOfAttacksRemaining <= 0 || !hasValidTargets() {
			cleanup()
			return nil
		}

		// For non-AoE attacks, handle single target logic
		if !aoe {
			monster, found := ctx.Data.Monsters.FindByID(settings.target)
			if !found || !isMonsterValid(monster) {
				cleanup()
				return nil
			}

			// TeleStomping
			if settings.telestomp && ctx.Data.CanTeleport() {
				if err := ensureEnemyIsInRange(monster, 2, 1); err != nil {
					cleanup()
					return err
				}
			}

			currentDistance := ctx.PathFinder.DistanceFromMe(monster.Position)

			// Handle ranged positioning
			if !settings.followEnemy && currentDistance > settings.maxDistance {
				if err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance); err != nil {
					ctx.Logger.Info("Enemy is out of range and cannot be reached",
						slog.Any("monster", monster.Name),
						slog.Int("distance", currentDistance))
					cleanup()
					return nil
				}
			}

			// Handle melee/following positioning
			if settings.followEnemy && (!settings.isMeleeAttack || lastRun.IsZero()) {
				if err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance); err != nil {
					ctx.Logger.Info("Enemy is out of range and cannot be reached",
						slog.Any("monster", monster.Name),
						slog.Int("distance", currentDistance))
					cleanup()
					return nil
				}
			}

			// Return to initial position if knocked back (melee only)
			if settings.isMeleeAttack && !lastRun.IsZero() {
				if distance := ctx.PathFinder.DistanceFromMe(monster.Position); distance > settings.maxDistance {
					if err := MoveTo(initialPosition); err != nil {
						ctx.Logger.Debug("Failed to return to initial position", slog.String("error", err.Error()))
					}
				}
			}
		}

		// Select correct skill for secondary attacks
		if !settings.primaryAttack && ctx.Data.PlayerUnit.RightSkill != settings.skill {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
			time.Sleep(time.Millisecond * 10)
		}

		// Handle aura activation
		if settings.aura != 0 && lastRun.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		// Attack timing check
		if time.Since(lastRun) <= ctx.Data.PlayerCastDuration()-attackCycleDuration || numOfAttacksRemaining <= 0 {
			continue
		}

		// Calculate attack target position
		x, y := 0, 0
		if aoe {
			var nearestDist float64 = 999999
			var nearestPos data.Position
			hasTarget := false

			for _, monster := range ctx.Data.Monsters.Enemies() {
				if !isMonsterValid(monster) {
					continue
				}

				distance := ctx.PathFinder.DistanceFromMe(monster.Position)
				if distance >= settings.minDistance && distance <= settings.maxDistance {
					if !hasTarget || float64(distance) < nearestDist {
						nearestDist = float64(distance)
						nearestPos = monster.Position
						hasTarget = true
					}
				}
			}

			if !hasTarget {
				cleanup()
				return nil
			}

			x, y = ctx.PathFinder.GameCoordsToScreenCords(nearestPos.X, nearestPos.Y)
		} else {
			monster, found := ctx.Data.Monsters.FindByID(settings.target)
			if !found || !isMonsterValid(monster) {
				cleanup()
				return nil
			}
			x, y = ctx.PathFinder.GameCoordsToScreenCords(monster.Position.X, monster.Position.Y)
		}

		// Perform the attack
		if settings.shouldStandStill {
			ctx.HID.KeyDown(ctx.Data.KeyBindings.StandStill)
		}

		if settings.isBurstCastSkill {
			ctx.HID.ReleaseMouseButton(game.RightButton)
		}

		if settings.primaryAttack {
			ctx.HID.Click(game.LeftButton, x, y)
		} else if settings.isBurstCastSkill {
			if hasValidTargets() {
				ctx.HID.HoldMouseButton(game.RightButton, x, y)
			}
		} else {
			ctx.HID.Click(game.RightButton, x, y)
		}

		if settings.shouldStandStill {
			ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
		}

		if !settings.isBurstCastSkill {
			ctx.HID.ReleaseMouseButton(game.RightButton)
		}

		lastRun = time.Now()
		numOfAttacksRemaining--

		// Handle burst skill cleanup
		if settings.isBurstCastSkill && (numOfAttacksRemaining <= 0 || !hasValidTargets()) {
			ctx.HID.ReleaseMouseButton(game.RightButton)
		}
	}
}

// ensureEnemyIsInRange handles positioning for attacks
func ensureEnemyIsInRange(monster data.Monster, maxDistance, minDistance int) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "ensureEnemyIsInRange"

	// Validate monster position before attempting any pathing
	if !ctx.Data.AreaData.IsInside(monster.Position) || !ctx.Data.AreaData.IsWalkable(monster.Position) {
		return errors.New("monster position is not valid or walkable")
	}

	path, distance, found := ctx.PathFinder.GetPath(monster.Position)

	// We cannot reach the enemy, let's skip the attack sequence
	if !found {
		return errors.New("path could not be calculated")
	}

	currentDistance := ctx.PathFinder.DistanceFromMe(monster.Position)
	hasLoS := ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, monster.Position)

	// We have line of sight and are within the valid range window, we can skip
	if hasLoS && currentDistance >= minDistance && currentDistance <= maxDistance {
		return nil
	}

	// In this case something weird is happening, just telestomp
	if distance < 2 {
		return MoveTo(monster.Position)
	}

	// If we're out of range, move to a safe position that respects max distance
	targetPos := ctx.PathFinder.GetSafePositionTowardsMonster(ctx.Data.PlayerUnit.Position, monster.Position, maxDistance)
	if targetPos != ctx.Data.PlayerUnit.Position {
		return MoveTo(targetPos)
	}

	// For monsters within range but without line of sight, try to find a position with LoS
	for i, pos := range path {
		distance = len(path) - i
		if distance > maxDistance || distance < minDistance {
			continue
		}

		if ctx.PathFinder.LineOfSight(pos, monster.Position) {
			safePos := ctx.PathFinder.GetSafePositionTowardsMonster(ctx.Data.PlayerUnit.Position, pos, maxDistance)
			return MoveTo(safePos)
		}
	}

	return nil
}
