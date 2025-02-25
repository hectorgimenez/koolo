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

	//TODO use the pathcache directly from path.go ?

	// Initialize or reuse path cache
	var pathCache *context.PathCache
	if ctx.CurrentGame.PathCache != nil &&
		ctx.CurrentGame.PathCache.DestPosition == dest &&
		IsPathValid(ctx.Data.PlayerUnit.Position, ctx.CurrentGame.PathCache) {
		// Reuse existing path cache
		pathCache = ctx.CurrentGame.PathCache
	} else {
		// Create new path cache
		start := ctx.Data.PlayerUnit.Position
		path, _, found := ctx.PathFinder.GetPath(dest)
		if !found {
			return fmt.Errorf("path not found to %v", dest)
		}

		pathCache = &context.PathCache{
			Path:             path,
			DestPosition:     dest,
			StartPosition:    start,
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

	for {
		time.Sleep(50 * time.Millisecond)
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		now := time.Now()

		// Refresh data and perform checks periodically
		if now.Sub(pathCache.LastCheck) > refreshInterval {
			ctx.RefreshGameData()
			currentPos := ctx.Data.PlayerUnit.Position
			distanceToDest := pather.DistanceFromPoint(currentPos, dest)

			// Check if we've reached destination
			if distanceToDest <= pathCache.DistanceToFinish || len(pathCache.Path) <= pathCache.DistanceToFinish {
				return nil
			}

			// Check for stuck in same position (direct equality check is efficient)
			isSamePosition := pathCache.PreviousPosition.X == currentPos.X && pathCache.PreviousPosition.Y == currentPos.Y

			// Only recalculate path in specific cases to reduce CPU usage
			if isSamePosition && !ctx.Data.CanTeleport() {
				// If stuck in same position without teleport, make random movement
				ctx.PathFinder.RandomMovement()
			} else if pathCache.Path == nil ||
				!IsPathValid(currentPos, pathCache) ||
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
				pathCache.Path = path
				pathCache.StartPosition = currentPos
			}

			pathCache.PreviousPosition = currentPos
			pathCache.LastCheck = now

			if now.Sub(startedAt) > timeout {
				return fmt.Errorf("movement timeout")
			}
		}

		if !ctx.Data.CanTeleport() {
			if time.Since(pathCache.LastRun) < walkDuration {
				continue
			}
		} else if time.Since(pathCache.LastRun) < ctx.Data.PlayerCastDuration() {
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

		pathCache.LastRun = time.Now()
		if len(pathCache.Path) > 0 {
			ctx.PathFinder.MoveThroughPath(pathCache.Path, walkDuration)
		}
	}
}

// Validate if the path is still valid based on current position
func IsPathValid(currentPos data.Position, cache *context.PathCache) bool {
	if cache == nil {
		return false
	}

	// Valid if we're close to start, destination, or current path
	if pather.DistanceFromPoint(currentPos, cache.StartPosition) < 20 ||
		pather.DistanceFromPoint(currentPos, cache.DestPosition) < 20 {
		return true
	}

	// Check if we're near any point on the path
	minDistance := 20
	for _, pathPoint := range cache.Path {
		dist := pather.DistanceFromPoint(currentPos, pathPoint)
		if dist < minDistance {
			minDistance = dist
			break
		}
	}

	return minDistance < 20
}
