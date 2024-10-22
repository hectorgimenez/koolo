package step

import (
	"errors"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const attackCycleDuration = 120 * time.Millisecond

type attackSettings struct {
	primaryAttack    bool
	skill            skill.ID
	followEnemy      bool
	minDistance      int
	maxDistance      int
	telestomp        bool
	aura             skill.ID
	target           data.UnitID
	shouldStandStill bool
	numOfAttacks     int
	isBurstCastSkill bool
}
type AttackOption func(step *attackSettings)

func Distance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = true
		step.minDistance = minimum
		step.maxDistance = maximum
	}
}

func EnsureAura(aura skill.ID) AttackOption {
	return func(step *attackSettings) {
		step.aura = aura
	}
}

func Telestomp() AttackOption {
	return func(step *attackSettings) {
		step.telestomp = true
	}
}

func PrimaryAttack(target data.UnitID, numOfAttacks int, standStill bool, opts ...AttackOption) error {
	ctx := context.Get()

	// Special case for Berserker
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

func SecondaryAttack(skill skill.ID, target data.UnitID, numOfAttacks int, opts ...AttackOption) error {
	settings := attackSettings{
		target:           target,
		numOfAttacks:     numOfAttacks,
		skill:            skill,
		primaryAttack:    false,
		isBurstCastSkill: skill == 48, //nova can define any other burst skill here
	}
	for _, o := range opts {
		o(&settings)
	}

	return attack(settings)
}

func attack(settings attackSettings) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "Attack"

	numOfAttacksRemaining := settings.numOfAttacks
	aoe := settings.target == 0
	lastRun := time.Time{}
	failedAttackAttempts := 0
	originalMaxDistance := settings.maxDistance
	var hasTargets bool
	var targetPos data.Position

	// Ensure keys/buttons are released when function exits or errors
	cleanup := func() {
		ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
		ctx.HID.ReleaseMouseButton(game.RightButton)
	}
	defer cleanup()

	// Helper function to check if there are any valid targets within range
	hasValidTargets := func(currentMaxDistance int) (bool, data.Position) {
		if !aoe {
			// For single target skills, just check the specific monster
			monster, found := ctx.Data.Monsters.FindByID(settings.target)
			if found && monster.Stats[stat.Life] > 0 {
				return true, monster.Position
			}
			return false, data.Position{}
		}

		// For AoE skills like Nova, check all monsters in range
		var nearestPos data.Position
		var nearestDist float64 = 999999
		hasTarget := false

		for _, monster := range ctx.Data.Monsters.Enemies() {
			distance := ctx.PathFinder.DistanceFromMe(monster.Position)
			if distance >= settings.minDistance && distance <= currentMaxDistance {
				if monster.Stats[stat.Life] > 0 {
					if !hasTarget || float64(distance) < nearestDist {
						nearestDist = float64(distance)
						nearestPos = monster.Position
						hasTarget = true
					}
				}
			}
		}
		return hasTarget, nearestPos
	}

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		// Adjust range if we're having trouble hitting the target
		currentMaxDistance := originalMaxDistance
		if failedAttackAttempts > 5 {
			if failedAttackAttempts == 6 {
				ctx.Logger.Debug("Looks like monster is not reachable, reducing max attack distance.")
			}
			currentMaxDistance = 1 // Reduce range to minimum when monsters seem unreachable
		}

		// Check for valid targets with current range
		hasTargets, targetPos = hasValidTargets(currentMaxDistance)
		if !hasTargets || numOfAttacksRemaining <= 0 {
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

			if settings.followEnemy {
				if err := ensureEnemyIsInRange(monster, currentMaxDistance, settings.minDistance); err != nil {
					failedAttackAttempts++
					if failedAttackAttempts > 10 {
						ctx.Logger.Info("Enemy remains unreachable after multiple attempts, aborting")
						cleanup()
						return nil
					}
					continue
				}
			}
		}

		// Ensure correct skill is selected for secondary attack
		if !settings.primaryAttack && ctx.Data.PlayerUnit.RightSkill != settings.skill {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
		}

		// Activate aura if necessary
		if settings.aura != 0 && lastRun.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		if time.Since(lastRun) > ctx.Data.PlayerCastDuration()-attackCycleDuration && numOfAttacksRemaining > 0 {
			x, y := 0, 0
			if aoe {
				x, y = ctx.PathFinder.GameCoordsToScreenCords(targetPos.X, targetPos.Y)
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

			// Perform attack
			if settings.primaryAttack {
				ctx.HID.Click(game.LeftButton, x, y)
			} else if settings.isBurstCastSkill {
				hasTargets, _ = hasValidTargets(currentMaxDistance)
				if hasTargets {
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
			failedAttackAttempts++ // Increment failed attempts - will be reset if we hit something
		}

		// For burst skills, check if we should release the button early
		if settings.isBurstCastSkill {
			hasTargets, _ = hasValidTargets(currentMaxDistance)
			if numOfAttacksRemaining <= 0 || !hasTargets {
				ctx.HID.ReleaseMouseButton(game.RightButton)
			}
		}

		// Check if we actually hit something - reset failed attempts if we did
		if lastRun.IsZero() {
			continue
		}
		hasTargets, _ = hasValidTargets(originalMaxDistance)
		if !hasTargets {
			failedAttackAttempts = 0 // Reset counter if we cleared the monsters
		}
	}
}
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
