package astar

import (
	"container/heap"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

var directions = []data.Position{
	{0, 1},   // Down
	{1, 0},   // Right
	{0, -1},  // Up
	{-1, 0},  // Left
	{1, 1},   // Down-Right (Southeast)
	{-1, 1},  // Down-Left (Southwest)
	{1, -1},  // Up-Right (Northeast)
	{-1, -1}, // Up-Left (Northwest)
}

type Node struct {
	data.Position
	Cost     int
	Priority int
	Index    int
}

func direction(from, to data.Position) (dx, dy int) {
	dx = to.X - from.X
	dy = to.Y - from.Y
	return
}

func CalculatePath(g *game.Grid, start, goal data.Position) ([]data.Position, int, bool) {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	costSoFar := make(map[data.Position]int)
	cameFrom := make(map[data.Position]data.Position)

	startNode := &Node{Position: start, Cost: 0, Priority: heuristic(start, goal)}
	heap.Push(&pq, startNode)
	costSoFar[start] = 0

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*Node)

		// Let's build the path if we reached the goal
		if current.Position == goal {
			var path []data.Position
			for p := goal; p != start; p = cameFrom[p] {
				path = append([]data.Position{p}, path...)
			}
			path = append([]data.Position{start}, path...)
			return path, len(path), true
		}

		for _, neighbor := range getNeighbors(g, current) {
			newCost := costSoFar[current.Position] + getCost(g.CollisionGrid[neighbor.Y][neighbor.X])

			// Handicap for changing direction, this prevents zig-zagging around obstacles
			curDirX, curDirY := direction(cameFrom[current.Position], current.Position)
			newDirX, newDirY := direction(current.Position, neighbor)
			if curDirX != newDirX || curDirY != newDirY {
				newCost += 2 // I found this value to be the acceptable
			}

			if _, ok := costSoFar[neighbor]; !ok || newCost < costSoFar[neighbor] {
				costSoFar[neighbor] = newCost
				priority := newCost + heuristic(neighbor, goal)
				heap.Push(&pq, &Node{Position: neighbor, Cost: newCost, Priority: priority})
				cameFrom[neighbor] = current.Position
			}
		}
	}

	return nil, 0, false
}

// Get walkable neighbors of a given node
func getNeighbors(grid *game.Grid, node *Node) []data.Position {
	var neighbors []data.Position

	for _, d := range directions {
		newPosition := data.Position{X: node.X + d.X, Y: node.Y + d.Y}
		if newPosition.X >= 0 && newPosition.X < grid.Width && newPosition.Y >= 0 && newPosition.Y < grid.Height {
			neighbors = append(neighbors, newPosition)
		}
	}

	return neighbors
}

func getCost(tileType game.CollisionType) int {
	switch tileType {
	case game.CollisionTypeWalkable:
		return 1 // Walkable
	case game.CollisionTypeMonster, game.CollisionTypeLowPriority:
		return 100 // Soft blocker
	default:
		return math.MaxInt32
	}
}

func heuristic(a, b data.Position) int {
	return int(math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y)))
}
