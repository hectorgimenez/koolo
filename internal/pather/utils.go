package pather

import (
	"math"
	"math/rand"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

func (pf *PathFinder) RandomMovement(d game.Data) {
	midGameX := pf.gr.GameAreaSizeX / 2
	midGameY := pf.gr.GameAreaSizeY / 2
	x := midGameX + rand.Intn(midGameX) - (midGameX / 2)
	y := midGameY + rand.Intn(midGameY) - (midGameY / 2)
	pf.hid.MovePointer(x, y)
	pf.hid.PressKeyBinding(d.KeyBindings.ForceMove)
	utils.Sleep(50)
}

func (pf *PathFinder) DistanceFromMe(p data.Position) int {
	return DistanceFromPoint(pf.data.PlayerUnit.Position, p)
}

func (pf *PathFinder) OptimizeRoomsTraverseOrder() []data.Room {
	distanceMatrix := make(map[data.Room]map[data.Room]int)

	for _, room1 := range pf.data.Rooms {
		distanceMatrix[room1] = make(map[data.Room]int)
		for _, room2 := range pf.data.Rooms {
			if room1 != room2 {
				_, distance, found := pf.GetClosestWalkablePathFrom(room1.GetCenter(), room2.GetCenter())
				if found {
					distanceMatrix[room1][room2] = distance
				} else {
					distanceMatrix[room1][room2] = math.MaxInt
				}
			} else {
				distanceMatrix[room1][room2] = 0
			}
		}
	}

	currentRoom := data.Room{}
	for _, r := range pf.data.Rooms {
		if r.IsInside(pf.data.PlayerUnit.Position) {
			currentRoom = r
		}
	}

	visited := make(map[data.Room]bool)
	order := []data.Room{currentRoom}
	visited[currentRoom] = true

	for len(order) < len(pf.data.Rooms) {
		nextRoom := data.Room{}
		minDistance := math.MaxInt

		// Find the nearest unvisited room
		for _, room := range pf.data.Rooms {
			if !visited[room] && distanceMatrix[currentRoom][room] < minDistance {
				nextRoom = room
				minDistance = distanceMatrix[currentRoom][room]
			}
		}

		// Add the next room to the order of visit
		order = append(order, nextRoom)
		visited[nextRoom] = true
		currentRoom = nextRoom
	}

	return order
}

func (pf *PathFinder) MoveThroughPath(p Path, distance int) {
	moveTo := p[len(p)-1]
	if distance > 0 && len(p) > distance {
		moveTo = p[distance]
	}

	screenX, screenY := pf.gameCoordsToScreenCords(p.From().X, p.From().Y, moveTo.X, moveTo.Y)
	// Prevent mouse overlap the HUD
	if screenY > int(float32(pf.gr.GameAreaSizeY)/1.21) {
		screenY = int(float32(pf.gr.GameAreaSizeY) / 1.21)
	}

	if distance > 0 {
		pf.MoveCharacter(screenX, screenY)
	}
}

func (pf *PathFinder) MoveCharacter(x, y int) {
	if pf.data.CanTeleport() {
		pf.hid.Click(game.RightButton, x, y)
	} else {
		pf.hid.MovePointer(x, y)
		pf.hid.PressKeyBinding(pf.data.KeyBindings.ForceMove)
		utils.Sleep(50)
	}
}

func (pf *PathFinder) GameCoordsToScreenCords(destinationX, destinationY int) (int, int) {
	return pf.gameCoordsToScreenCords(pf.data.PlayerUnit.Position.X, pf.data.PlayerUnit.Position.Y, destinationX, destinationY)
}

func (pf *PathFinder) gameCoordsToScreenCords(playerX, playerY, destinationX, destinationY int) (int, int) {
	// Calculate diff between current player position and destination
	diffX := destinationX - playerX
	diffY := destinationY - playerY

	// Transform cartesian movement (World) to isometric (screen)
	// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
	screenX := int((float32(diffX-diffY) * 19.8) + float32(pf.gr.GameAreaSizeX/2))
	screenY := int((float32(diffX+diffY) * 9.9) + float32(pf.gr.GameAreaSizeY/2))

	return screenX, screenY
}

func IsNarrowMap(a area.ID) bool {
	switch a {
	case area.MaggotLairLevel1, area.MaggotLairLevel2, area.MaggotLairLevel3, area.ArcaneSanctuary, area.ClawViperTempleLevel2, area.RiverOfFlame, area.ChaosSanctuary:
		return true
	}

	return false
}

func DistanceFromPoint(from data.Position, to data.Position) int {
	first := math.Pow(float64(to.X-from.X), 2)
	second := math.Pow(float64(to.Y-from.Y), 2)

	return int(math.Sqrt(first + second))
}

func (pf *PathFinder) LineOfSight(origin data.Position, destination data.Position) bool {
	x0, y0 := origin.X-pf.data.AreaOrigin.X, origin.Y-pf.data.AreaOrigin.Y
	x1, y1 := destination.X-pf.data.AreaOrigin.X, destination.Y-pf.data.AreaOrigin.Y

	dx := int(math.Abs(float64(x1 - x0)))
	dy := int(math.Abs(float64(y1 - y0)))
	sx := -1
	sy := -1

	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}

	err := dx - dy

	for {
		// Boundary check for x0, y0
		if x0 < 0 || y0 < 0 || x0 >= len(pf.data.AreaData.Grid.CollisionGrid[0]) || y0 >= len(pf.data.AreaData.Grid.CollisionGrid) {
			return false
		}

		// Check if the current position is not walkable
		if !pf.data.AreaData.Grid.CollisionGrid[y0][x0] {
			return false
		}

		// Check if we have reached the end point
		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}

	return true
}