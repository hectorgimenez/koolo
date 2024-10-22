package step

import (
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/mode"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func MoveTo(dest data.Position) error {
	ctx := context.Get()
	ctx.ContextDebug.LastStep = "MoveTo"

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
	stopAtDistance := 7
	idleThreshold := time.Second * 3
	idleStartTime := time.Time{}

	startedAt := time.Now()
	lastRun := time.Time{}
	previousPosition := data.Position{}

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
			if ctx.PathFinder.DistanceFromMe(dest) < stopAtDistance+5 {
				return nil
			}
			ctx.Logger.Error("Path could not be calculated",
				slog.Any("destination", dest),
				slog.Any("player_position", ctx.Data.PlayerUnit.Position),
				slog.String("area", ctx.Data.PlayerUnit.Area.Area().Name))
			return errors.New("path could not be calculated, maybe there is an obstacle or a flying platform (arcane sanctuary)")
		}
		if distance <= stopAtDistance || len(path) <= stopAtDistance || len(path) == 0 {
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

		previousPosition = ctx.Data.PlayerUnit.Position
		ctx.PathFinder.MoveThroughPath(path, walkDuration)
	}
}
func GetSafePositionTowardsMonster(playerPos, monsterPos data.Position, safeDistance int) data.Position {
	dx := float64(monsterPos.X - playerPos.X)
	dy := float64(monsterPos.Y - playerPos.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > float64(safeDistance) {
		ratio := float64(safeDistance) / distance
		safePos := data.Position{
			X: playerPos.X + int(dx*ratio),
			Y: playerPos.Y + int(dy*ratio),
		}
		return FindNearestWalkablePosition(safePos)
	}

	return playerPos
}

func GetSafePositionAwayFromMonster(playerPos, monsterPos data.Position, safeDistance int) data.Position {
	dx := float64(playerPos.X - monsterPos.X)
	dy := float64(playerPos.Y - monsterPos.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance < float64(safeDistance) {
		ratio := float64(safeDistance) / distance
		safePos := data.Position{
			X: monsterPos.X + int(dx*ratio),
			Y: monsterPos.Y + int(dy*ratio),
		}
		return FindNearestWalkablePosition(safePos)
	}

	return playerPos
}

// FindNearestWalkablePosition finds the nearest walkable position to the given position
func FindNearestWalkablePosition(pos data.Position) data.Position {
	ctx := context.Get()
	if ctx.Data.AreaData.Grid.IsWalkable(pos) {
		return pos
	}

	for radius := 1; radius <= 10; radius++ {
		for x := pos.X - radius; x <= pos.X+radius; x++ {
			for y := pos.Y - radius; y <= pos.Y+radius; y++ {
				checkPos := data.Position{X: x, Y: y}
				if ctx.Data.AreaData.Grid.IsWalkable(checkPos) {
					return checkPos
				}
			}
		}
	}

	return pos
}
