package step

import (
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
			if !ensureEnemyIsInRange(monster, 2, 1) {
				path, _, _ := ctx.PathFinder.GetClosestWalkablePath(monster.Position)
				if len(path) > 0 {
					// Move to the closest tile to the monster
					MoveTo(path[len(path)-1])
				}
			}
		}

		if !aoe {
			// Move into the attack distance range before starting
			if settings.followEnemy {
				if !ensureEnemyIsInRange(monster, settings.maxDistance, settings.minDistance) {
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
func ensureEnemyIsInRange(monster data.Monster, maxDistance, minDistance int) bool {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "ensureEnemyIsInRange"

	_, distance, found := ctx.PathFinder.GetPath(monster.Position)

	// We cannot reach the enemy, let's skip the attack sequence
	if !found {
		return false
	}

	hasLoS := ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, monster.Position)

	if distance > maxDistance || !hasLoS {
		if distance > minDistance && minDistance > 0 {
			// TODO tweak this crap
			//moveTo := minDistance - 1
			//if len(path) < minDistance {
			//	moveTo = len(path) - 1 // Ensure moveTo is within path bounds
			//}

			MoveTo(data.Position{X: monster.Position.X, Y: monster.Position.Y})
			//for i := moveTo; i > 0; i-- {
			//	posTile := path[i].(*pather.Tile)
			//	pos := data.Position{
			//		X: posTile.X + ctx.Data.AreaOrigin.X,
			//		Y: posTile.Y + ctx.Data.AreaOrigin.Y,
			//	}
			//
			//	hasLoS = ctx.PathFinder.LineOfSight(pos, monster.Position)
			//	if hasLoS {
			//		path, distance, _ = ctx.PathFinder.GetPath(pos)
			//		_ = MoveTo(pos)
			//	}
			//}
		}
	}

	return true
}
