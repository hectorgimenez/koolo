package step

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/pather"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const DistanceToFinishMoving = 4

type MoveOption func(*context.PathCache)

// WithDistanceToFinish overrides the default DistanceToFinishMoving
func WithDistanceToFinish(distance int) MoveOption {
	return func(cache *context.PathCache) {
		cache.DistanceToFinish = distance
	}
}

// TODO town pathing stucks?

func MoveTo(dest data.Position, options ...MoveOption) error {
	ctx := context.Get()
	ctx.SetLastStep("MoveTo")

	//Todo if we add this it will make walkable character move like a broken disk. However is helping with waypoint interaction
	// reducing walkduration seems to help with this issue.

	/*	defer func() {
		for {
			switch ctx.Data.PlayerUnit.Mode {
			case mode.Walking, mode.WalkingInTown, mode.Running, mode.CastingSkill:
				utils.Sleep(50)
				//	ctx.RefreshGameData()
				continue
			default:
				return
			}
		}
	}()*/

	const (
		refreshInterval = 200 * time.Millisecond
		timeout         = 30 * time.Second
	)

	startedAt := time.Now()

	// Initialize or reuse path cache
	var pathCache *context.PathCache
	currentPos := ctx.Data.PlayerUnit.Position
	currentArea := ctx.Data.PlayerUnit.Area
	canTeleport := ctx.Data.CanTeleport()

	if ctx.CurrentGame.PathCache != nil &&
		ctx.CurrentGame.PathCache.DestPosition == dest &&
		ctx.CurrentGame.PathCache.IsPathValid(currentPos) {
		// Reuse existing path cache
		pathCache = ctx.CurrentGame.PathCache
	} else {
		// Get path from global cache or calculate new one
		path, found := pather.GetCachedPath(currentPos, dest, currentArea, canTeleport)
		if !found {
			// Calculate new path if not found in cache
			path, _, found = ctx.PathFinder.GetPath(dest)
			if !found {
				return fmt.Errorf("path not found to %v", dest)
			}
			// Store in global cache
			pather.StorePath(currentPos, dest, currentArea, canTeleport, path, currentPos)
		}

		// Create new PathCache reference
		pathCache = &context.PathCache{
			Path:             path,
			DestPosition:     dest,
			StartPosition:    currentPos,
			DistanceToFinish: DistanceToFinishMoving,
		}
		ctx.CurrentGame.PathCache = pathCache
	}

	// Apply any options
	for _, option := range options {
		option(pathCache)
	}

	// Add some delay between clicks to let the character move to destination
	walkDuration := utils.RandomDurationMs(700, 900)

	// Get last check time from global cache
	lastCheck := pathCache.GetLastCheck(currentArea, canTeleport)
	lastRun := pathCache.GetLastRun(currentArea, canTeleport)
	previousPosition := pathCache.GetPreviousPosition(currentArea, canTeleport)

	for {
		time.Sleep(50 * time.Millisecond)
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		now := time.Now()

		// Refresh data and perform checks periodically
		if now.Sub(lastCheck) > refreshInterval {
			ctx.RefreshGameData()
			currentPos = ctx.Data.PlayerUnit.Position
			distanceToDest := pather.DistanceFromPoint(currentPos, dest)

			// Check if we've reached destination
			if distanceToDest <= pathCache.DistanceToFinish || len(pathCache.Path) <= pathCache.DistanceToFinish {
				return nil
			}

			// Check for stuck in same position (direct equality check is efficient)
			isSamePosition := previousPosition.X == currentPos.X && previousPosition.Y == currentPos.Y

			// Only recalculate path in specific cases to reduce CPU usage
			if isSamePosition && !ctx.Data.CanTeleport() {
				// If stuck in same position without teleport, make random movement
				ctx.PathFinder.RandomMovement()
			} else if pathCache.Path == nil ||
				!pathCache.IsPathValid(currentPos) ||
				(distanceToDest <= 15 && distanceToDest > pathCache.DistanceToFinish) {
				//TODO this looks like the telestomp issue,  IsSamePosition is true but it never enter this condition, need something else to force refresh

				// Only recalculate when truly needed
				path, _, found := ctx.PathFinder.GetPath(dest)
				if !found {
					if ctx.PathFinder.DistanceFromMe(dest) < pathCache.DistanceToFinish+5 {
						return nil // Close enough
					}
					return fmt.Errorf("failed to calculate path")
				}

				// Update global cache and local PathCache
				pather.StorePath(currentPos, dest, currentArea, canTeleport, path, currentPos)
				pathCache.Path = path
				pathCache.StartPosition = currentPos
			}

			// Update previous position and check time
			previousPosition = currentPos
			lastCheck = now
			pathCache.UpdateLastCheck(currentArea, canTeleport, currentPos)

			if now.Sub(startedAt) > timeout {
				return fmt.Errorf("movement timeout")
			}
		}

		if !ctx.Data.CanTeleport() {
			if time.Since(lastRun) < walkDuration {
				continue
			}
		} else if time.Since(lastRun) < ctx.Data.PlayerCastDuration() {
			continue
		}

		// Press the Teleport keybinding if it's available, otherwise use vigor (if available)
		if ctx.Data.CanTeleport() {
			walkDuration = ctx.Data.PlayerCastDuration()
			if ctx.Data.PlayerUnit.RightSkill != skill.Teleport {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.Teleport))
				time.Sleep(50 * time.Millisecond)
				continue
			}
		} else if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Vigor); found {
			if ctx.Data.PlayerUnit.RightSkill != skill.Vigor {
				ctx.HID.PressKeyBinding(kb)
				time.Sleep(50 * time.Millisecond)
				continue
			}
		}

		lastRun = time.Now()
		pathCache.UpdateLastRun(currentArea, canTeleport)

		if len(pathCache.Path) > 0 {
			ctx.PathFinder.MoveThroughPath(pathCache.Path, walkDuration)
		}
	}
}
