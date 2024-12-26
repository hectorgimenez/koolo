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

func MoveTo(dest data.Position) error {
	minDistanceToFinishMoving := DistanceToFinishMoving

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
		ctx.RefreshGameData()

		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

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

		// Add some delay between clicks to let the character move to destination
		walkDuration := utils.RandomDurationMs(600, 1200)
		if !ctx.Data.CanTeleport() && time.Since(lastRun) < walkDuration {
			continue
		}

		// We skip the movement if we can teleport and the last movement time was less than the player cast duration
		if ctx.Data.CanTeleport() && time.Since(lastRun) < ctx.Data.PlayerCastDuration() {
			continue
		}

		lastRun = time.Now()

		// If we are stuck in the same position, make a random movement and cross fingers
		if previousPosition == ctx.Data.PlayerUnit.Position && !ctx.Data.CanTeleport() {
			ctx.PathFinder.RandomMovement()
			continue
		}

		// This is a workaround to avoid the character to get stuck in the same position when the hitbox of the destination is too big
		if distance < 20 && math.Abs(float64(previousDistance-distance)) < 4 {
			minDistanceToFinishMoving += 4
		} else {
			minDistanceToFinishMoving = DistanceToFinishMoving
		}

		previousPosition = ctx.Data.PlayerUnit.Position
		previousDistance = distance
		ctx.PathFinder.MoveThroughPath(path, walkDuration)
	}
}
