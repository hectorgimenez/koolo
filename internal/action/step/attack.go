package step

import (
	"errors"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

const attackCycleDuration = 120 * time.Millisecond

type attackSettings struct {
	primaryAttack     bool
	skill             skill.ID
	followEnemy       bool
	minDistance       int
	maxDistance       int
	telestomp         bool
	aura              skill.ID
	target            data.UnitID
	shouldStandStill  bool
	numOfAttacks      int
	isBurstCastSkill  bool
	isMeleeAttack     bool
	currentlyBursting bool
}

type AttackOption func(step *attackSettings)

func Distance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = true
		step.minDistance = minimum
		step.maxDistance = maximum
		step.isMeleeAttack = minimum <= 1 && maximum <= 3
	}
}

func RangedDistance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = false
		step.minDistance = minimum
		step.maxDistance = maximum
		step.isMeleeAttack = false
	}
}

func StationaryDistance(minimum, maximum int) AttackOption {
	return func(step *attackSettings) {
		step.followEnemy = false
		step.minDistance = minimum
		step.maxDistance = maximum
		step.isMeleeAttack = false
		step.shouldStandStill = true
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
		isBurstCastSkill: skill == 48, // nova
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
	lastMovement := time.Time{}
	initialPosition := ctx.Data.PlayerUnit.Position
	isAttackInProgress := false

	cleanup := func() {
		ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
		ctx.HID.ReleaseMouseButton(game.RightButton)
		settings.currentlyBursting = false
		isAttackInProgress = false
	}
	defer cleanup()

	hasValidTargets := func() bool {
		if !aoe {
			monster, found := ctx.Data.Monsters.FindByID(settings.target)
			return found && monster.Stats[stat.Life] > 0
		}

		for _, monster := range ctx.Data.Monsters.Enemies() {
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
		ctx.PauseIfNotPriority()

		if numOfAttacksRemaining <= 0 || !hasValidTargets() {
			cleanup()
			return nil
		}

		allowMovement := !settings.currentlyBursting &&
			!isAttackInProgress &&
			time.Since(lastRun) > ctx.Data.PlayerCastDuration() &&
			time.Since(lastMovement) > time.Millisecond*200

		if allowMovement {
			if !aoe {
				monster, found := ctx.Data.Monsters.FindByID(settings.target)
				if !found || monster.Stats[stat.Life] <= 0 {
					cleanup()
					return nil
				}

				currentDistance := ctx.PathFinder.DistanceFromMe(monster.Position)
				hasLoS := ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, monster.Position)

				needsRepositioning := currentDistance > settings.maxDistance ||
					currentDistance < settings.minDistance ||
					!hasLoS

				if needsRepositioning {
					if err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance, settings); err != nil {
						cleanup()
						return err
					}
					lastMovement = time.Now()
					// Add delay after movement
					time.Sleep(time.Millisecond * 100)
				}

				// Handle melee repositioning
				if settings.isMeleeAttack && !lastRun.IsZero() {
					if distance := ctx.PathFinder.DistanceFromMe(monster.Position); distance > settings.maxDistance {
						if err := MoveTo(initialPosition); err != nil {
							ctx.Logger.Debug("Failed to return to initial position", slog.String("error", err.Error()))
						}
						lastMovement = time.Now()
					}
				}
			}
		}

		if !settings.primaryAttack && ctx.Data.PlayerUnit.RightSkill != settings.skill && !settings.currentlyBursting {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
			time.Sleep(time.Millisecond * 40)
		}

		if settings.aura != 0 && lastRun.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		if time.Since(lastRun) > ctx.Data.PlayerCastDuration()-attackCycleDuration && numOfAttacksRemaining > 0 {
			isAttackInProgress = true

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

			if settings.shouldStandStill {
				ctx.HID.KeyDown(ctx.Data.KeyBindings.StandStill)
			}

			if settings.primaryAttack {
				ctx.HID.Click(game.LeftButton, x, y)
			} else if settings.isBurstCastSkill {
				if !settings.currentlyBursting {
					ctx.HID.HoldMouseButton(game.RightButton, x, y)
					settings.currentlyBursting = true
				}
			} else {
				ctx.HID.Click(game.RightButton, x, y)
			}

			if settings.shouldStandStill && !settings.isBurstCastSkill {
				ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
			}

			lastRun = time.Now()
			numOfAttacksRemaining--

			if !settings.isBurstCastSkill {
				time.Sleep(ctx.Data.PlayerCastDuration())
				isAttackInProgress = false
			}
		}

		if settings.currentlyBursting && !hasValidTargets() {
			ctx.HID.ReleaseMouseButton(game.RightButton)
			if settings.shouldStandStill {
				ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
			}
			settings.currentlyBursting = false
			isAttackInProgress = false
		}
	}
}

func ensureEnemyIsInRange(monster data.Monster, maxDistance, minDistance int, settings attackSettings) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "ensureEnemyIsInRange"

	path, distance, found := ctx.PathFinder.GetPath(monster.Position)
	if !found {
		return errors.New("path could not be calculated")
	}

	hasLoS := ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, monster.Position)

	if hasLoS && distance < maxDistance {
		return nil
	}

	// For Mosaic Sin stack building - stay on monster
	if maxDistance == 0 && minDistance == 0 {
		if distance < 2 {
			return MoveTo(monster.Position)
		}
		return nil
	}
	// First release stand still if it's active
	if settings.shouldStandStill {
		ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
		time.Sleep(40 * time.Millisecond) // Small delay to ensure key release
	}
	// Handle ranged characters (like FoH)
	if !settings.isBurstCastSkill && minDistance > 3 {
		currentDistance := ctx.PathFinder.DistanceFromMe(monster.Position)
		if currentDistance > maxDistance || !hasLoS {
			for _, pos := range path {
				potentialPosition := data.Position{
					X: pos.X + ctx.Data.AreaOrigin.X,
					Y: pos.Y + ctx.Data.AreaOrigin.Y,
				}
				distanceToMonster := pather.DistanceFromPoint(potentialPosition, monster.Position)
				if distanceToMonster <= maxDistance && ctx.PathFinder.LineOfSight(potentialPosition, monster.Position) {
					utils.Sleep(100)
					return MoveTo(potentialPosition)
				}
			}
		}
		return nil
	}
	// For any other characters just move directly to monster for now
	for i, pos := range path {
		distance = len(path) - i
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
