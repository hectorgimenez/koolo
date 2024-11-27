package action

import (
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func ClearCurrentLevel(openChests bool, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.SetLastAction("ClearCurrentLevel")

	rooms := ctx.PathFinder.OptimizeRoomsTraverseOrder()
	for _, r := range rooms {
		err := clearRoom(r, filter)
		if err != nil {
			ctx.Logger.Warn("Failed to clear room: %v", err)
		}

		if !openChests {
			continue
		}

		for _, o := range ctx.Data.Objects {
			if o.IsChest() && o.Selectable && r.IsInside(o.Position) {
				err = MoveToCoords(o.Position)
				if err != nil {
					ctx.Logger.Warn("Failed moving to chest: %v", err)
					continue
				}
				err = InteractObject(o, func() bool {
					chest, _ := ctx.Data.Objects.FindByID(o.ID)
					return !chest.Selectable
				})
				if err != nil {
					ctx.Logger.Warn("Failed interacting with chest: %v", err)
				}
				utils.Sleep(500) // Add small delay to allow the game to open the chest and drop the content
			}
		}
	}

	return nil
}

func clearRoom(room data.Room, filter data.MonsterFilter) error {
	ctx := context.Get()
	ctx.SetLastAction("clearRoom")

	// Get points and filter only for basic validity
	validPoints := make([]data.Position, 0)
	for _, p := range getClearPoints(&room) {
		if ctx.Data.AreaData.IsInside(p) && ctx.Data.AreaData.IsWalkable(p) {
			validPoints = append(validPoints, p)
		}
	}

	// Sort points by simple distance from player
	sort.Slice(validPoints, func(i, j int) bool {
		distI := ctx.PathFinder.DistanceFromMe(validPoints[i])
		distJ := ctx.PathFinder.DistanceFromMe(validPoints[j])
		return distI < distJ
	})

	// Try to move to any valid point
	for _, point := range validPoints {
		if err := MoveToCoords(point); err == nil {
			goto ClearMonsters
		}
	}

	// If we couldn't move to any point, try to clear from current position

ClearMonsters:

	for {
		monsters := getMonstersInRoom(room, filter)
		if len(monsters) == 0 {
			return nil
		}

		sort.Slice(monsters, func(i, j int) bool {
			if monsters[i].IsMonsterRaiser() != monsters[j].IsMonsterRaiser() {
				return monsters[i].IsMonsterRaiser()
			}
			distI := ctx.PathFinder.DistanceFromMe(monsters[i].Position)
			distJ := ctx.PathFinder.DistanceFromMe(monsters[j].Position)
			return distI < distJ
		})

		for _, monster := range monsters {
			if !ctx.Data.AreaData.IsInside(monster.Position) {
				continue
			}

			if !ctx.PathFinder.LineOfSight(ctx.Data.PlayerUnit.Position, monster.Position) {
				// Let MoveToCoords handle the pathing (it uses GetPath)
				if err := MoveToCoords(monster.Position); err != nil {
					continue
				}
			}

			ctx.Char.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
				m, found := d.Monsters.FindByID(monster.UnitID)
				if found && m.Stats[stat.Life] > 0 {
					return monster.UnitID, true
				}
				return 0, false
			}, nil)
		}
	}
}

func getClearPoints(room *data.Room) []data.Position {
	points := make([]data.Position, 0)

	center := room.GetCenter()
	points = append(points, center)

	edges := []struct{ x, y float64 }{
		{0.2, 0.2}, {0.8, 0.2}, // Top edge
		{0.2, 0.8}, {0.8, 0.8}, // Bottom edge
		{0.5, 0.5},             // Middle
		{0.5, 0.2}, {0.5, 0.8}, // Additional middle points
		{0.2, 0.5}, {0.8, 0.5},
	}

	for _, edge := range edges {
		points = append(points, data.Position{
			X: room.Position.X + int(float64(room.Width)*edge.x),
			Y: room.Position.Y + int(float64(room.Height)*edge.y),
		})
	}

	return points
}

func getMonstersInRoom(room data.Room, filter data.MonsterFilter) []data.Monster {
	ctx := context.Get()
	ctx.SetLastAction("getMonstersInRoom")

	monstersInRoom := make([]data.Monster, 0)
	for _, m := range ctx.Data.Monsters.Enemies(filter) {
		// Skip dead monsters or those outside current area
		if m.Stats[stat.Life] <= 0 || !ctx.Data.AreaData.IsInside(m.Position) {
			continue
		}

		inRoom := room.IsInside(m.Position)
		nearPlayer := ctx.PathFinder.DistanceFromMe(m.Position) < 30
		nearRoom := false

		if !inRoom && !nearPlayer {
			roomCenter := room.GetCenter()
			distToRoom := pather.DistanceFromPoint(roomCenter, m.Position)
			nearRoom = distToRoom < room.Width/2+15 || distToRoom < room.Height/2+15
		}

		if inRoom || nearPlayer || nearRoom {
			monstersInRoom = append(monstersInRoom, m)
		}
	}

	return monstersInRoom
}
