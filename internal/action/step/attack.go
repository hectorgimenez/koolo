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
		target:       target,
		numOfAttacks: numOfAttacks,
		skill:        skill,
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

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		monster, found := ctx.Data.Monsters.FindByID(settings.target)
		if !found || monster.Stats[stat.Life] <= 0 || numOfAttacksRemaining <= 0 {
			return nil
		}

		// TeleStomp
		if settings.telestomp && ctx.Data.CanTeleport() {
			err := ensureEnemyIsInRange(monster, 2, 1)
			if err != nil {
				return err
			}
		}

		if !aoe {
			// Move into the attack distance range before starting
			if settings.followEnemy {
				if err := ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance); err != nil {
					// We cannot reach the enemy, let's skip the attack sequence
					ctx.Logger.Info("Enemy is out of range and can not be reached", slog.Any("monster", monster.Name))
					return nil
				}
			} else {
				// Since we are not following the enemy, and it's not in range, we can't attack it
				_, distance, found := ctx.PathFinder.GetPath(monster.Position)
				if !found || distance > settings.maxDistance {
					return nil
				}
			}
		}

		// If we are not using the primary attack, we need to ensure the right skill is selected
		if !settings.primaryAttack && ctx.Data.PlayerUnit.RightSkill != settings.skill {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.skill))
			time.Sleep(time.Millisecond * 80)
		}

		// If we have an aura, let's ensure it's active
		if settings.aura != 0 && lastRun.IsZero() {
			ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(settings.aura))
		}

		if time.Since(lastRun) > ctx.Data.PlayerCastDuration()-attackCycleDuration && numOfAttacksRemaining > 0 {
			if settings.shouldStandStill {
				ctx.HID.KeyDown(ctx.Data.KeyBindings.StandStill)
			}
			x, y := ctx.PathFinder.GameCoordsToScreenCords(monster.Position.X, monster.Position.Y)

			if settings.primaryAttack {
				ctx.HID.Click(game.LeftButton, x, y)
			} else {
				ctx.HID.Click(game.RightButton, x, y)
			}
			if settings.shouldStandStill {
				ctx.HID.KeyUp(ctx.Data.KeyBindings.StandStill)
			}
			lastRun = time.Now()
			numOfAttacksRemaining--
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
