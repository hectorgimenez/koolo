package step

import (
	"errors"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"log/slog"
	"time"

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

// attack performs the main attack sequence based on provided settings
func attack(settings attackSettings) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "Attack"

	numOfAttacksRemaining := settings.numOfAttacks
	aoe := settings.target == 0
	lastRun := time.Time{}
	initialPosition := ctx.Data.PlayerUnit.Position

	// Ensure keys/buttons are released when function exits or errors
	cleanup := func() {
		standStillKey := ctx.Data.KeyBindings.StandStill.Key1[0]
		if ctx.HID.IsKeyPressed(standStillKey) {
			ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
		}
		// Release any held mouse buttons
		ctx.HID.ReleaseMouseButton(game.RightButton)
	}
	defer cleanup() // Ensure cleanup happens when function exit

	// Skip monsters that are off grid unless it's a special case
	isMonsterValid := func(monster data.Monster) bool {
		// Special case: Always allow Vizier seal boss even if off grid
		isVizier := monster.Type == data.MonsterTypeSuperUnique && monster.Name == npc.StormCaster
		if !isVizier && !ctx.Data.AreaData.IsInside(monster.Position) {
			return false
		}
		return monster.Stats[stat.Life] > 0
	}
	// Check if we have any valid targets within range
	hasValidTargets := func() bool {
		if !aoe {
			// For single target skills, just check the specific monster
			monster, found := ctx.Data.Monsters.FindByID(settings.target)
			return found && monster.Stats[stat.Life] > 0
		}

		// For AoE skills like Nova, check all monsters in range
		for _, monster := range ctx.Data.Monsters.Enemies() {
			if !isMonsterValid(monster) {
				continue
			}
			distance := ctx.PathFinder.DistanceFromMe(monster.Position)
			if distance >= settings.minDistance && distance <= settings.maxDistance {
				if monster.Stats[stat.Life] > 0 {
					return true
				}
			}
		}
		return false
	}

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		// Check if we should continue attacking based on remaining attacks and valid targets
		if numOfAttacksRemaining <= 0 || !hasValidTargets() {
			cleanup()
			return nil
		}

		// For non-AoE attacks, handle telestomp and range checks
		if !aoe {
			monster, found := ctx.Data.Monsters.FindByID(settings.target)
			if !found || monster.Stats[stat.Life] <= 0 {
				cleanup()
				return nil
			}

			// TeleStomp
			if settings.telestomp && ctx.Data.CanTeleport() {
				if err := ensureEnemyIsInRange(monster, 2, 1); err != nil {
					cleanup()
					return err
				}
			}

			currentDistance := ctx.PathFinder.DistanceFromMe(monster.Position)

			// Handle ranged attacks positioning
			if !settings.followEnemy && currentDistance > settings.maxDistance {
				if err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance); err != nil {
					ctx.Logger.Info("Enemy is out of range and cannot be reached", slog.Any("monster", monster.Name))
					cleanup()
					return nil
				}
			}

			// Handle melee/following attacks positioning
			if settings.followEnemy && (!settings.isMeleeAttack || lastRun.IsZero()) {
				if err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance); err != nil {
					ctx.Logger.Info("Enemy is out of range and cannot be reached", slog.Any("monster", monster.Name))
					cleanup()
					return nil
				}
			}

			// For melee attacks, return to initial position if we've been knocked back significantly
			if settings.isMeleeAttack && !lastRun.IsZero() {
				if distance := ctx.PathFinder.DistanceFromMe(monster.Position); distance > settings.maxDistance {
					if err := MoveTo(initialPosition); err != nil {
						ctx.Logger.Debug("Failed to return to initial position", slog.String("error", err.Error()))
					}
				}
			}
		}

		// Ensure correct skill is selected for secondary attack
		if !settings.primaryAttack && ctx.Data.PlayerUnit.RightSkill != settings.skill {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
			time.Sleep(time.Millisecond * 10)
		}

		// Activate aura if necessary
		if settings.aura != 0 && lastRun.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		// Only attack if enough time has passed or starting a new burst
		if time.Since(lastRun) > ctx.Data.PlayerCastDuration()-attackCycleDuration && numOfAttacksRemaining > 0 {

			// Calculate target position
			x, y := 0, 0
			if aoe {
				var nearestDist float64 = 999999
				var nearestPos data.Position
				hasTarget := false

				for _, monster := range ctx.Data.Monsters.Enemies() {
					distance := ctx.PathFinder.DistanceFromMe(monster.Position)
					if distance >= settings.minDistance && distance <= settings.maxDistance && monster.Stats[stat.Life] > 0 {
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
				monster, _ := ctx.Data.Monsters.FindByID(settings.target)
				x, y = ctx.PathFinder.GameCoordsToScreenCords(monster.Position.X, monster.Position.Y)
			}

			// Press StandStill if required
			if settings.shouldStandStill {
				ctx.HID.KeyDown(ctx.Data.KeyBindings.StandStill)
			}

			// For burst skills, release any previously held right click before starting new attack
			if settings.isBurstCastSkill {
				ctx.HID.ReleaseMouseButton(game.RightButton)
			}

			// Perform the actual attack
			if settings.primaryAttack {
				ctx.HID.Click(game.LeftButton, x, y)
			} else if settings.isBurstCastSkill {
				if hasValidTargets() {
					ctx.HID.HoldMouseButton(game.RightButton, x, y)
				}
			} else {
				ctx.HID.Click(game.RightButton, x, y)
			}

			// Release StandStill immediately after attack
			if settings.shouldStandStill {
				ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
			}

			// Release right mouse button for non-burst skills
			if !settings.isBurstCastSkill {
				ctx.HID.ReleaseMouseButton(game.RightButton)
			}

			lastRun = time.Now()
			numOfAttacksRemaining--
		}

		// For burst skills, check if we should release the button early
		if settings.isBurstCastSkill {
			if numOfAttacksRemaining <= 0 || !hasValidTargets() {
				// End burst if no valid targets remain
				ctx.HID.ReleaseMouseButton(game.RightButton)
			}
		}
	}
}

// ensureEnemyIsInRange handles positioning for attacks
func ensureEnemyIsInRange(monster data.Monster, maxDistance, minDistance int) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "ensureEnemyIsInRange"

	path, distance, found := ctx.PathFinder.GetPath(monster.Position)

	// We cannot reach the enemy, let's skip the attack sequence
	if !found {
		return errors.New("path could not be calculated")
	}

	hasLoS := ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, monster.Position)

	// We have line of sight, and we are inside the attack range, we can skip
	if hasLoS && distance < maxDistance {
		return nil
	}

	for i, pos := range path {
		distance = len(path) - i
		// In this case something weird is happening, just telestomp
		if distance < 2 {
			return MoveTo(monster.Position)
		}

		if distance > maxDistance {
			continue
		}

		if ctx.PathFinder.LineOfSight(pos, monster.Position) {
			return MoveTo(pos)
		}
	}

	return nil
}
