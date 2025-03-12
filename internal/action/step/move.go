package step

import (
	"fmt"
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

const (
	DistanceToFinishMoving = 4
	idleThreshold          = 1500 * time.Millisecond // Changed from 3s to 1.5s
	maxMovementTimeout     = 30 * time.Second
)

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
	ctx := context.Get()
	ctx.SetLastStep("MoveTo")

	opts := &MoveOpts{}
	for _, o := range options {
		o(opts)
	}

	minDistanceToFinishMoving := DistanceToFinishMoving
	if opts.distanceOverride != nil {
		minDistanceToFinishMoving = *opts.distanceOverride
	}

	startedAt := time.Now()
	lastRun := time.Time{}
	previousPosition := data.Position{}
	previousDistance := 0
	idleStartTime := time.Time{}

	for {
		ctx.PauseIfNotPriority()
		ctx.RefreshGameData()

		// Timeout check
		if time.Since(startedAt) > maxMovementTimeout {
			return nil
		}

		// Distance check
		currentDistance := ctx.PathFinder.DistanceFromMe(dest)
		if currentDistance <= minDistanceToFinishMoving {
			return nil
		}

		// Idle detection
		if ctx.Data.PlayerUnit.Position == previousPosition {
			if idleStartTime.IsZero() {
				idleStartTime = time.Now()
			} else if time.Since(idleStartTime) > idleThreshold {
				ctx.Logger.Debug("Anti-idle triggered")
				ctx.PathFinder.RandomMovement()
				idleStartTime = time.Time{}
				time.Sleep(150 * time.Millisecond) // Reduced from 300ms
				continue
			}
		} else {
			idleStartTime = time.Time{}
			previousPosition = ctx.Data.PlayerUnit.Position
		}

		// Skill management
		if ctx.Data.CanTeleport() {
			if ctx.Data.PlayerUnit.RightSkill != skill.Teleport {
				ctx.HID.PressKeyBinding(ctx.Data.KeyBindings.MustKBForSkill(skill.Teleport))
			}
		} else if kb, found := ctx.Data.KeyBindings.KeyBindingForSkill(skill.Vigor); found {
			if ctx.Data.PlayerUnit.RightSkill != skill.Vigor {
				ctx.HID.PressKeyBinding(kb)
			}
		}

		// Path calculation
		path, distance, found := ctx.PathFinder.GetPath(dest)
		if !found {
			if currentDistance < minDistanceToFinishMoving+5 {
				return nil
			}
			return fmt.Errorf("pathfinding failed in %s to %d,%d",
				ctx.Data.PlayerUnit.Area.Area().Name,
				dest.X,
				dest.Y,
			)
		}

		// Dynamic distance adjustment
		if distance < 20 && math.Abs(float64(previousDistance-distance)) < DistanceToFinishMoving {
			minDistanceToFinishMoving += DistanceToFinishMoving
		} else if opts.distanceOverride != nil {
			minDistanceToFinishMoving = *opts.distanceOverride
		} else {
			minDistanceToFinishMoving = DistanceToFinishMoving
		}

		// Movement execution
		walkDuration := utils.RandomDurationMs(600, 1200)
		if !ctx.Data.CanTeleport() && time.Since(lastRun) < walkDuration {
			time.Sleep(walkDuration - time.Since(lastRun))
			continue
		}

		if ctx.Data.CanTeleport() && time.Since(lastRun) < ctx.Data.PlayerCastDuration() {
			// Calculate 70% of the original cast duration
			fasterCastDuration := ctx.Data.PlayerCastDuration() * 7 / 10 // 70% of original duration | higher FCR -> 5 / 10 (50% speed = 2x faster)
			remainingWait := fasterCastDuration - time.Since(lastRun)

			if remainingWait > 0 {
				time.Sleep(remainingWait)
			}
			continue
		}

		ctx.PathFinder.MoveThroughPath(path, walkDuration)
		lastRun = time.Now()
		previousDistance = distance
	}
}
