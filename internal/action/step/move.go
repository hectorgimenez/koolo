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
	"github.com/hectorgimenez/koolo/internal/pather"
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
	lastPathCheck := time.Time{}
	var cachedPath pather.Path
	var cachedDistance int

	for {
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()
		// is needed to prevent bot teleporting in circle when it reached destination (lower end cpu) cost is minimal.
		ctx.RefreshGameData()

		// Check for idle state outside town
		if ctx.Data.PlayerUnit.Mode == mode.StandingOutsideTown {
			if idleStartTime.IsZero() {
				idleStartTime = time.Now()
			} else if time.Since(idleStartTime) > idleThreshold {
				// Perform anti-idle action
				ctx.Logger.Debug("Anti-idle triggered")
				ctx.PathFinder.RandomMovement()
				idleStartTime = time.Time{} // Reset idle timer
				continue
			}
		} else {
			idleStartTime = time.Time{} // Reset idle timer if not in StandingOutsideTown mode
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

		// Path checking to reduce CPU load
		var path pather.Path
		var distance int
		var found bool
		if time.Since(lastPathCheck) > 100*time.Millisecond {
			path, distance, found = ctx.PathFinder.GetPath(dest)
			
			// Attempt progressive teleport steps when path is blocked
			if !found && ctx.Data.CanTeleport() {
				if intermediatePath, _, intermediateFound := ctx.PathFinder.GetClosestWalkablePath(dest); intermediateFound {
					if len(intermediatePath) > 0 {
						intermediatePos := intermediatePath[len(intermediatePath)-1]
						if ctx.PathFinder.DistanceFromMe(intermediatePos) < 15 {
							ctx.Logger.Debug("Taking intermediate teleport step to bypass obstacle")
							if err := MoveTo(intermediatePos, WithDistanceToFinish(2)); err == nil {
								continue // Retry original destination after intermediate move
							}
						}
					}
				}
			}
			
			lastPathCheck = time.Now()
			cachedPath = path
			cachedDistance = distance
		} else {
			path = cachedPath
			distance = cachedDistance
			found = path != nil
		}

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
		
		// Add some delay between clicks to let the character move to destination
		walkDuration := utils.RandomDurationMs(600, 1200)
		if !ctx.Data.CanTeleport() && time.Since(lastRun) < walkDuration {
			utils.Sleep(50)
			continue
		}

		// We skip the movement if we can teleport and the last movement time was less than the player cast duration
		if ctx.Data.CanTeleport() && time.Since(lastRun) < ctx.Data.PlayerCastDuration() {
			utils.Sleep(50)
			continue
		}

		lastRun = time.Now()

		// If we are stuck in the same position, make a random movement and cross fingers
		if previousPosition == ctx.Data.PlayerUnit.Position && !ctx.Data.CanTeleport() {
			ctx.PathFinder.RandomMovement()
			continue
		}

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

		// Reduce CPU usage of loop iterations
		utils.Sleep(50)
	}
}
