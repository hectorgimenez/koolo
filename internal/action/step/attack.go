package step

import (
	"errors"
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

	// Ensure keys/buttons are released when function exits or errors
	cleanup := func() {
		ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
		ctx.HID.ReleaseMouseButton(game.RightButton)
	}
	defer cleanup()

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		monster, found := ctx.Data.Monsters.FindByID(settings.target)
		if !found || monster.Stats[stat.Life] <= 0 || numOfAttacksRemaining <= 0 {
			cleanup() // Explicitly cleanup before returning
			return nil
		}

		// TeleStomp
		if settings.telestomp && ctx.Data.CanTeleport() {
			if err := ensureEnemyIsInRange(monster, 2, 1); err != nil {
				cleanup() // Explicitly cleanup before returning error
				return err
			}
		}

		if !aoe && settings.followEnemy {
			if err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance); err != nil {
				ctx.Logger.Info("Enemy is out of range and can not be reached", slog.Any("monster", monster.Name))
				cleanup() // Explicitly cleanup before returning
				return nil
			}
		}

		// Ensure correct skill is selected for secondary attack
		if !settings.primaryAttack && ctx.Data.PlayerUnit.RightSkill != settings.skill {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
			time.Sleep(time.Millisecond * 80)
		}

		// Activate aura if necessary
		if settings.aura != 0 && lastRun.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		if time.Since(lastRun) > ctx.Data.PlayerCastDuration()-attackCycleDuration && numOfAttacksRemaining > 0 {
			x, y := ctx.PathFinder.GameCoordsToScreenCords(monster.Position.X, monster.Position.Y)

			// Press StandStill if required
			if settings.shouldStandStill {
				ctx.HID.KeyDown(ctx.Data.KeyBindings.StandStill)
			}

			// For burst skills, release any previously held right click before starting new attack
			if settings.isBurstCastSkill {
				ctx.HID.ReleaseMouseButton(game.RightButton)
				//		time.Sleep(time.Millisecond * 50) // Small delay to ensure button is fully released
			}

			// Perform attack
			if settings.primaryAttack {
				ctx.HID.Click(game.LeftButton, x, y)
			} else if settings.isBurstCastSkill {
				// Only hold the button if target is still valid
				if monster.Stats[stat.Life] > 0 {
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
			// Release if target is dead or we're done with attacks
			if numOfAttacksRemaining <= 0 || monster.Stats[stat.Life] <= 0 {
				ctx.HID.ReleaseMouseButton(game.RightButton)
				//		time.Sleep(time.Millisecond * 50) // Small delay to ensure button is fully released
			}
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
