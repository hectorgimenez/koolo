package step

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const DistanceToFinishMoving = 4

type MoveOpts struct {
	distanceOverride *int
}

type MoveOption func(*MoveOpts)

// WithDistanceToFinish overrides the default DistanceToFinishMoving
func WithDistanceToFinish(distance int) MoveOption {
	return func(opts *MoveOpts) {
		opts.distanceOverride = &distance
	}
}

func MoveTo(dest data.Position, options ...MoveOption) error {
	// Initialize options
	opts := &MoveOpts{}

	// Apply any provided options
	for _, o := range options {
		o(opts)
	}

	minDistanceToFinishMoving := DistanceToFinishMoving
	if opts.distanceOverride != nil {
		minDistanceToFinishMoving = *opts.distanceOverride
	}

	ctx := context.Get()
	ctx.SetLastStep("MoveTo")

	defer func() {
		for {
			switch ctx.Data.PlayerUnit.Mode {
			case mode.Walking, mode.WalkingInTown, mode.Running, mode.CastingSkill:
				utils.Sleep(100)
				ctx.RefreshGameData()
				continue
			default:
				return
			}
		}
	}()

	timeout := time.Second * 30
	idleThreshold := time.Second * 3
	idleStartTime := time.Time{}

	startedAt := time.Now()
	lastRun := time.Time{}
	previousPosition := data.Position{}
	previousDistance := 0

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()
		// is needed to prevent bot teleporting in circle when it reached destination (lower end cpu) cost is minimal.
		ctx.RefreshGameData()

		// Add some delay between clicks to let the character move to destination
		walkDuration := utils.RandomDurationMs(600, 1200)
		if !ctx.Data.CanTeleport() && time.Since(lastRun) < walkDuration {
			time.Sleep(walkDuration - time.Since(lastRun))
			continue
		}

		// We skip the movement if we can teleport and the last movement time was less than the player cast duration
		if ctx.Data.CanTeleport() && time.Since(lastRun) < ctx.Data.PlayerCastDuration() {
			time.Sleep(ctx.Data.PlayerCastDuration() - time.Since(lastRun))
			continue
		}

		// Check for idle state
		if ctx.Data.PlayerUnit.Position == previousPosition {
			if idleStartTime.IsZero() {
				idleStartTime = time.Now()
			} else if time.Since(idleStartTime) > idleThreshold {
				// Perform anti-idle action
				ctx.Logger.Debug("Anti-idle triggered")

				if ctx.CharacterCfg.Character.UseTeleport {
					ctx.PathFinder.RandomTeleport()
				} else {
					ctx.PathFinder.RandomMovement()
				}

				idleStartTime = time.Time{} // Reset idle timer
				continue
			}
		} else {
			idleStartTime = time.Time{} // Reset idle timer if position changed
			previousPosition = ctx.Data.PlayerUnit.Position
		}

		// Press the Teleport keybinding if it's available, otherwise use vigor (if available)
		if ctx.Data.CanTeleport() {
			if ctx.Data.PlayerUnit.RightSkill != skill.Teleport {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.Teleport))
			}
		} else if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Vigor); found {
			if ctx.Data.PlayerUnit.RightSkill != skill.Vigor {
				ctx.HID.PressKeyBinding(kb)
			}
		}

		path, distance, found := ctx.PathFinder.GetPath(dest)
		if !found {
			if ctx.PathFinder.DistanceFromMe(dest) < minDistanceToFinishMoving+5 {
				return nil
			}

			return errors.New("path could not be calculated. Current area: [" + ctx.Data.PlayerUnit.Area.Area().Name + "]. Trying to path to Destination: [" + fmt.Sprintf("%d,%d", dest.X, dest.Y) + "]")
		}
		if distance <= minDistanceToFinishMoving || len(path) <= minDistanceToFinishMoving || len(path) == 0 {
			return nil
		}

		// Exit on timeout
		if timeout > 0 && time.Since(startedAt) > timeout {
			return nil
		}

		lastRun = time.Now()

		// This is a workaround to avoid the character to get stuck in the same position when the hitbox of the destination is too big
		if distance < 20 && math.Abs(float64(previousDistance-distance)) < DistanceToFinishMoving {
			minDistanceToFinishMoving += DistanceToFinishMoving
		} else if opts.distanceOverride != nil {
			minDistanceToFinishMoving = *opts.distanceOverride
		} else {
			minDistanceToFinishMoving = DistanceToFinishMoving
		}

		previousPosition = ctx.Data.PlayerUnit.Position
		previousDistance = distance
		ctx.PathFinder.MoveThroughPath(path, walkDuration)
	}
}
