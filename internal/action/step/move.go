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

func MoveTo(dest data.Position, options ...MoveOption) error {
	ctx := context.Get()
	ctx.SetLastStep("MoveTo")

	const (
		refreshInterval = 200 * time.Millisecond
		timeout         = 30 * time.Second
	)

	startedAt := time.Now()
	minDistanceToFinishMoving := DistanceToFinishMoving
	//previousDistance := -1

	if ctx.CurrentGame.PathCache != nil {
		if pather.DistanceFromPoint(dest, ctx.CurrentGame.PathCache.DestPosition) > 2 {
			ctx.InvalidatePathCache("destination changed")
		}
	}
	// Initialize or reuse path cache
	var pathCache *context.PathCache
	if ctx.CurrentGame.PathCache != nil && IsPathValid(ctx.Data.PlayerUnit.Position, ctx.CurrentGame.PathCache) {
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
			LastCheck:        time.Time{},
			LastRun:          time.Time{},
			PreviousPosition: data.Position{},
		}
		ctx.CurrentGame.PathCache = pathCache
	}

	// Apply any options
	for _, option := range options {
		option(pathCache)
	}

	// Add some delay between clicks to let the character move to destination
	walkDuration := utils.RandomDurationMs(700, 1100)

	for {
		time.Sleep(50 * time.Millisecond)
		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

		now := time.Now()

		// Refresh data and perform checks periodically
		if now.Sub(pathCache.LastCheck) > refreshInterval {
			ctx.RefreshGameData()
			currentPos := ctx.Data.PlayerUnit.Position
			distance := pather.DistanceFromPoint(currentPos, dest)

			// Simple stuck detection - if position hasn't changed, handle appropriately
			if pathCache.PreviousPosition.X == currentPos.X && pathCache.PreviousPosition.Y == currentPos.Y &&
				pathCache.PreviousPosition.X != 0 && pathCache.PreviousPosition.Y != 0 {

				if !ctx.Data.CanTeleport() {
					// For non-teleporters, immediately make a random movement
					ctx.PathFinder.RandomMovement()
					ctx.InvalidatePathCache("stuck walking")
				} else {
					// For teleporters, only invalidate if we've been stuck for multiple checks
					// Check if this is the first time noticing we're stuck
					if pathCache.StuckCount == 0 {
						// First time we notice being stuck, just increment counter
						pathCache.StuckCount++
					} else if pathCache.StuckCount >= 2 {
						// Only invalidate after being stuck for 3 consecutive checks
						ctx.InvalidatePathCache("stuck teleporting")
						pathCache.StuckCount = 0

						// Recalculate path
						path, _, found := ctx.PathFinder.GetPath(dest)
						if !found {
							if ctx.PathFinder.DistanceFromMe(dest) < minDistanceToFinishMoving+5 {
								return nil // Close enough
							}
							return fmt.Errorf("failed to calculate path")
						}
						pathCache.Path = path
						pathCache.StartPosition = currentPos
					} else {
						pathCache.StuckCount++
					}
				}
			} else {
				// Reset stuck counter when position changes
				pathCache.StuckCount = 0
			}

			/*	// This is a workaround to avoid the character getting stuck when the hitbox of the destination is too big
				if distance < 20 && previousDistance != -1 && math.Abs(float64(previousDistance-distance)) < DistanceToFinishMoving {
					minDistanceToFinishMoving += DistanceToFinishMoving
				} else {
					minDistanceToFinishMoving = pathCache.DistanceToFinish
				}*/

			//		previousDistance = distance
			pathCache.PreviousPosition = currentPos

			// Check if we've reached destination
			if distance <= minDistanceToFinishMoving || len(pathCache.Path) <= minDistanceToFinishMoving {
				return nil
			}

			if pathCache.Path == nil ||
				!IsPathValid(currentPos, pathCache) {
				// Only recalculate when truly needed
				path, _, found := ctx.PathFinder.GetPath(dest)
				if !found {
					if ctx.PathFinder.DistanceFromMe(dest) < minDistanceToFinishMoving+5 {
						return nil // Close enough
					}
					return fmt.Errorf("failed to calculate path")
				}
				pathCache.Path = path
				pathCache.StartPosition = currentPos
			}

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
	ctx := context.Get()

	if cache == nil || len(cache.Path) == 0 {
		return false
	}

	// Convert current position to relative coordinates (grid-relative)
	// Path points are stored in grid-relative coordinates (without area offset)
	relativeCurrentPos := data.Position{
		X: currentPos.X - ctx.Data.AreaOrigin.X,
		Y: currentPos.Y - ctx.Data.AreaOrigin.Y,
	}

	// Check if we're near any point on the path (path points are in relative coordinates)
	minDistance := 3
	for _, pathPoint := range cache.Path {
		dist := pather.DistanceFromPoint(relativeCurrentPos, pathPoint)
		if dist < minDistance {
			minDistance = dist
		}
	}

	return minDistance < 3
}
