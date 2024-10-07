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

	// Use a 2D slice to store the cost of each node
	costSoFar := make([][]int, g.Width)
	cameFrom := make([][]data.Position, g.Width)
	for i := range costSoFar {
		costSoFar[i] = make([]int, g.Height)
		cameFrom[i] = make([]data.Position, g.Height)
		for j := range costSoFar[i] {
			costSoFar[i][j] = math.MaxInt32
		}
	}

	startNode := &Node{Position: start, Cost: 0, Priority: heuristic(start, goal)}
	heap.Push(&pq, startNode)
	costSoFar[start.X][start.Y] = 0

	neighbors := make([]data.Position, 0, 8)

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*Node)

		// Let's build the path if we reached the goal
		if current.Position == goal {
			var path []data.Position
			for p := goal; p != start; p = cameFrom[p.X][p.Y] {
				path = append([]data.Position{p}, path...)
			}
			path = append([]data.Position{start}, path...)
			return path, len(path), true
		}

		updateNeighbors(g, current, &neighbors)

		for _, neighbor := range neighbors {
			newCost := costSoFar[current.X][current.Y] + getCost(g.CollisionGrid[neighbor.Y][neighbor.X])

			// Handicap for changing direction, this prevents zig-zagging around obstacles
			//curDirX, curDirY := direction(cameFrom[current.X][current.Y], current.Position)
			//newDirX, newDirY := direction(current.Position, neighbor)
			//if curDirX != newDirX || curDirY != newDirY {
			//	newCost++
			//}

			if newCost < costSoFar[neighbor.X][neighbor.Y] {
				costSoFar[neighbor.X][neighbor.Y] = newCost
				priority := newCost + int(0.5*float64(heuristic(neighbor, goal)))
				heap.Push(&pq, &Node{Position: neighbor, Cost: newCost, Priority: priority})
				cameFrom[neighbor.X][neighbor.Y] = current.Position
			}
		}
	}

	return nil, 0, false
}

// Get walkable neighbors of a given node
func updateNeighbors(grid *game.Grid, node *Node, neighbors *[]data.Position) {
	*neighbors = (*neighbors)[:0]

	x, y := node.X, node.Y
	gridWidth, gridHeight := grid.Width, grid.Height

	for _, d := range directions {
		newX, newY := x+d.X, y+d.Y
		// Check if the new neighbor is within grid bounds
		if newX >= 0 && newX < gridWidth && newY >= 0 && newY < gridHeight {
			*neighbors = append(*neighbors, data.Position{X: newX, Y: newY})
		}
	}
}

func getCost(tileType game.CollisionType) int {
	switch tileType {
	case game.CollisionTypeWalkable:
		return 1 // Walkable
	case game.CollisionTypeMonster:
		return 16
	case game.CollisionTypeLowPriority, game.CollisionTypeObject:
		return 4 // Soft blocker
	default:
		return math.MaxInt32
	}
}

func heuristic(a, b data.Position) int {
	dx := math.Abs(float64(a.X - b.X))
	dy := math.Abs(float64(a.Y - b.Y))
	return int(dx + dy + (math.Sqrt(2)-2)*math.Min(dx, dy))
}
