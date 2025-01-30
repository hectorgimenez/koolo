package astar

import (
	"container/heap"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/game"
)

var (
	cardinalDirections = []data.Position{
		{0, 1},  // Down
		{1, 0},  // Right
		{0, -1}, // Up
		{-1, 0}, // Left
	}
	allDirections = []data.Position{
		{0, 1},   // Down
		{1, 0},   // Right
		{0, -1},  // Up
		{-1, 0},  // Left
		{1, 1},   // Down-Right
		{-1, 1},  // Down-Left
		{1, -1},  // Up-Right
		{-1, -1}, // Up-Left
	}
)

type Node struct {
	data.Position
	Cost     int
	Priority int
	Index    int
}

func CalculatePath(g *game.Grid, areaID area.ID, start, goal data.Position) ([]data.Position, int, bool) {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

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
	nodesExplored := 0

	directions := allDirections
	if IsNarrowMap(areaID) {
		directions = cardinalDirections
	}

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*Node)
		nodesExplored++

		if current.Position == goal {
			return reconstructPath(cameFrom, start, goal), nodesExplored, true
		}

		updateNeighbors(g, current, directions, &neighbors)
		for _, neighbor := range neighbors {
			tileCost := getCost(g.CollisionGrid[neighbor.Y][neighbor.X])
			if tileCost == math.MaxInt32 {
				continue // Skip completely blocked tiles
			}

			newCost := costSoFar[current.X][current.Y] + tileCost
			if newCost < costSoFar[neighbor.X][neighbor.Y] {
				costSoFar[neighbor.X][neighbor.Y] = newCost
				priority := newCost + heuristic(neighbor, goal)
				heap.Push(&pq, &Node{Position: neighbor, Cost: newCost, Priority: priority})
				cameFrom[neighbor.X][neighbor.Y] = current.Position
			}
		}
	}

	return nil, 0, false
}

func reconstructPath(cameFrom [][]data.Position, start, goal data.Position) []data.Position {
	var path []data.Position
	for p := goal; p != start; p = cameFrom[p.X][p.Y] {
		path = append([]data.Position{p}, path...)
	}
	return append([]data.Position{start}, path...)
}

func updateNeighbors(grid *game.Grid, node *Node, directions []data.Position, neighbors *[]data.Position) {
	*neighbors = (*neighbors)[:0]
	x, y := node.X, node.Y

	for _, d := range directions {
		newX, newY := x+d.X, y+d.Y
		if newX >= 0 && newX < grid.Width && newY >= 0 && newY < grid.Height {
			tileType := grid.CollisionGrid[newY][newX]
			if tileType == game.CollisionTypeNonWalkable {
				continue // Skip non-walkable tiles
			}
			*neighbors = append(*neighbors, data.Position{X: newX, Y: newY})
		}
	}
}

var tileCost = map[game.CollisionType]int{
	game.CollisionTypeWalkable:    1,
	game.CollisionTypeMonster:     16,
	game.CollisionTypeObject:      4,
	game.CollisionTypeLowPriority: 20,
	game.CollisionTypeNonWalkable: math.MaxInt32, // Completely block non-walkable
}

func getCost(tileType game.CollisionType) int {
	return tileCost[tileType]
}

func heuristic(a, b data.Position) int {
	dx := abs(a.X - b.X)
	dy := abs(a.Y - b.Y)
	return dx + dy
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func IsNarrowMap(a area.ID) bool {
	switch a {
	case area.MaggotLairLevel1, area.MaggotLairLevel2, area.MaggotLairLevel3, area.ArcaneSanctuary, area.ClawViperTempleLevel2, area.RiverOfFlame, area.ChaosSanctuary:
		return true
	}
	return false
}
