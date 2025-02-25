package pather

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/utils"
)

var (
	// Cached walkable positions with mutex protection
	walkablePosCache = make(map[string]data.Position)
	walkablePosLock  sync.RWMutex
)

func (pf *PathFinder) RandomMovement() {
	midGameX := pf.gr.GameAreaSizeX / 2
	midGameY := pf.gr.GameAreaSizeY / 2
	x := midGameX + rand.Intn(midGameX) - (midGameX / 2)
	y := midGameY + rand.Intn(midGameY) - (midGameY / 2)
	pf.hid.MovePointer(x, y)
	pf.hid.PressKeyBinding(pf.data.KeyBindings.ForceMove)
	utils.Sleep(50)
}

func (pf *PathFinder) DistanceFromMe(p data.Position) int {
	return DistanceFromPoint(pf.data.PlayerUnit.Position, p)
}

// Search in expanding squares around target position with caching
func (pf *PathFinder) FindNearbyWalkablePosition(target data.Position) (data.Position, bool) {
	key := fmt.Sprintf("%d:%d", target.X, target.Y)
	// Check cache first with read lock
	walkablePosLock.RLock()
	if pos, exists := walkablePosCache[key]; exists {
		walkablePosLock.RUnlock()
		return pos, true
	}

	walkablePosLock.RUnlock()

	// Search in expanding squares around target position
	for radius := 1; radius <= 3; radius++ {
		for x := -radius; x <= radius; x++ {
			for y := -radius; y <= radius; y++ {
				if x == 0 && y == 0 {
					continue // Skip center position
				}
				pos := data.Position{X: target.X + x, Y: target.Y + y}
				if pf.data.AreaData.IsWalkable(pos) {
					// Update cache with write lock
					walkablePosLock.Lock()
					walkablePosCache[key] = pos
					walkablePosLock.Unlock()
					return pos, true
				}
			}
		}
	}
	return data.Position{}, false
}

// Create optimal room visiting order using nearest neighbor algorithm
func (pf *PathFinder) OptimizeRoomsTraverseOrder() []data.Room {
	distanceMatrix := make(map[data.Room]map[data.Room]int)

	// Build distance matrix between all rooms
	for _, room1 := range pf.data.Rooms {
		distanceMatrix[room1] = make(map[data.Room]int)
		for _, room2 := range pf.data.Rooms {
			if room1 != room2 {
				distance := DistanceFromPoint(room1.GetCenter(), room2.GetCenter())
				distanceMatrix[room1][room2] = distance
			} else {
				distanceMatrix[room1][room2] = 0
			}
		}
	}

	// Find current room based on player position
	currentRoom := data.Room{}
	for _, r := range pf.data.Rooms {
		if r.IsInside(pf.data.PlayerUnit.Position) {
			currentRoom = r
		}
	}

	visited := make(map[data.Room]bool)
	order := []data.Room{currentRoom}
	visited[currentRoom] = true

	// Nearest neighbor pathfinding
	for len(order) < len(pf.data.Rooms) {
		nextRoom := data.Room{}
		minDistance := math.MaxInt

		for _, room := range pf.data.Rooms {
			if !visited[room] && distanceMatrix[currentRoom][room] < minDistance {
				nextRoom = room
				minDistance = distanceMatrix[currentRoom][room]
			}
		}

		// Find closest unvisited room
		order = append(order, nextRoom)
		visited[nextRoom] = true
		currentRoom = nextRoom
	}

	return order
}

// Navigate along a path considering movement constraints
func (pf *PathFinder) MoveThroughPath(p Path, walkDuration time.Duration) {
	// Calculate the max distance we can walk in the given duration
	maxDistance := int(float64(25) * walkDuration.Seconds())

	// Let's try to calculate how close to the window border we can go
	screenCords := data.Position{}
	for distance, pos := range p {
		screenX, screenY := pf.gameCoordsToScreenCords(p.From().X, p.From().Y, pos.X, pos.Y)

		// We reached max distance, let's stop (if we are not teleporting)
		if !pf.data.CanTeleport() && maxDistance > 0 && distance > maxDistance {
			break
		}

		// Prevent mouse overlap the HUD
		if screenY > int(float32(pf.gr.GameAreaSizeY)/1.21) {
			break
		}

		// We are getting out of the window, let's stop
		if screenX < 0 || screenY < 0 || screenX > pf.gr.GameAreaSizeX || screenY > pf.gr.GameAreaSizeY {
			break
		}
		screenCords = data.Position{X: screenX, Y: screenY}
	}

	pf.MoveCharacter(screenCords.X, screenCords.Y)
}

// Handle movement based on character capabilities
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

// Identify areas requiring restricted movement directions
func IsNarrowMap(a area.ID) bool {
	switch a {
	case area.MaggotLairLevel1, area.MaggotLairLevel2, area.MaggotLairLevel3,
		area.ArcaneSanctuary, area.ClawViperTempleLevel2, area.RiverOfFlame,
		area.ChaosSanctuary:
		return true
	}
	return false
}

// Calculate straight-line distance between two positions (Bresenham algo)
func DistanceFromPoint(from data.Position, to data.Position) int {
	first := math.Pow(float64(to.X-from.X), 2)
	second := math.Pow(float64(to.Y-from.Y), 2)

	return int(math.Sqrt(first + second))
}

// Check if there's unobstructed path between two points
func (pf *PathFinder) LineOfSight(origin data.Position, destination data.Position) bool {
	dx := int(math.Abs(float64(destination.X - origin.X)))
	dy := int(math.Abs(float64(destination.Y - origin.Y)))
	sx, sy := 1, 1

	if origin.X > destination.X {
		sx = -1
	}
	if origin.Y > destination.Y {
		sy = -1
	}

	err := dx - dy

	x, y := origin.X, origin.Y

	for {
		if !pf.data.AreaData.Grid.IsWalkable(data.Position{X: x, Y: y}) {
			return false
		}
		if x == destination.X && y == destination.Y {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}

	return true
}

// Calculate a new position that is a specified distance beyond the target position when viewed from the start position (calculates position extended beyond target point)
func (pf *PathFinder) BeyondPosition(start, target data.Position, distance int) data.Position {
	// Calculate direction vector
	dx := float64(target.X - start.X)
	dy := float64(target.Y - start.Y)

	// Normalize
	length := math.Sqrt(dx*dx + dy*dy)
	if length == 0 {
		// If positions are identical, pick arbitrary direction
		dx = 1
		dy = 0
	} else {
		dx = dx / length
		dy = dy / length
	}

	// Return position extended beyond target
	return data.Position{
		X: target.X + int(dx*float64(distance)),
		Y: target.Y + int(dy*float64(distance)),
	}
}
