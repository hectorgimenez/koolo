package step

import (
	"errors"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/v2/context"
	"github.com/hectorgimenez/koolo/internal/v2/utils"
)

func MoveTo(dest data.Position) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "MoveTo"

	timeout := time.Second * 30
	stopAtDistance := 7

	startedAt := time.Now()
	lastRun := time.Time{}

	for {
		ctx.RefreshGameData()

		// Pause the execution if the priority is not the same as the execution priority
		ctx.PauseIfNotPriority()

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

		distance := ctx.PathFinder.DistanceFromMe(dest)
		if distance <= stopAtDistance {
			// In case distance is lower, we double-check with the pathfinder and the full path instead of euclidean distance
			_, distance, found := ctx.PathFinder.GetPath(dest)
			if !found || distance <= stopAtDistance {
				return nil
			}
		}

		if timeout > 0 && time.Since(startedAt) > timeout {
			return nil
		}

		// Add some delay between clicks to let the character move to destination
		walkDuration := utils.RandomDurationMs(600, 1200)
		if !ctx.Data.CanTeleport() && time.Since(lastRun) < walkDuration {
			continue
		}

		if ctx.Data.CanTeleport() && time.Since(lastRun) < ctx.Data.PlayerCastDuration() {
			continue
		}

		// TODO Implement stuck & cache?

		path, _, found := ctx.PathFinder.GetClosestWalkablePath(dest)
		if !found {
			if ctx.PathFinder.DistanceFromMe(dest) < stopAtDistance+5 {
				return nil
			}

			return errors.New("path could not be calculated, maybe there is an obstacle or a flying platform (arcane sanctuary)")
		}
		lastRun = time.Now()
		if len(path) == 0 {
			return nil
		}
		//lastRunPositions = append(m.lastRunPositions, d.PlayerUnit.Position)
		ctx.PathFinder.MoveThroughPath(path, calculateMaxDistance(ctx.Data, walkDuration))
	}
}

func calculateMaxDistance(d *game.Data, duration time.Duration) int {
	// We don't care too much if teleport is available, we can ignore corners, 90 degrees turns, etc
	if d.CanTeleport() {
		return 25
	}

	// Calculate the distance we can walk in the given duration, based on the randomized time
	proposedDistance := int(float64(25) * duration.Seconds())
	realDistance := proposedDistance

	return realDistance
}
