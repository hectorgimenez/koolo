package astar

import (
	"container/heap"
	"math"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Node struct {
	data.Position
	G, H, F float64
	Parent  *Node
}

func CalculatePath(g *game.Grid, start, goal data.Position) ([]data.Position, float64, bool) {
	openList := &PriorityQueue{}
	heap.Init(openList)

	startNode := &Node{Position: start, G: 0, H: heuristic(start, goal)}
	startNode.F = startNode.G + startNode.H
	heap.Push(openList, startNode)

	closedSet := make(map[data.Position]bool)
	openListMap := make(map[data.Position]bool)
	openListMap[start] = true

	for openList.Len() > 0 {
		currentNode := heap.Pop(openList).(*Node)
		delete(openListMap, currentNode.Position)
		closedSet[currentNode.Position] = true

		if currentNode.Position == goal {
			return reconstructPath(currentNode), currentNode.G, true
		}

		for _, neighbor := range getNeighbors(g, currentNode) {
			if closedSet[neighbor.Position] {
				continue
			}

			tentativeG := currentNode.G + 1
			if tentativeG < neighbor.G || neighbor.G == 0 {
				neighbor.G = tentativeG
				neighbor.H = heuristic(neighbor.Position, goal)
				neighbor.F = neighbor.G + neighbor.H
				neighbor.Parent = currentNode

				if _, inOpenList := openListMap[neighbor.Position]; !inOpenList {
					heap.Push(openList, neighbor)
					openListMap[neighbor.Position] = true
				}
			}
		}
	}

	return nil, 0, false
}

// Get walkable neighbors of a given node
func getNeighbors(grid *game.Grid, node *Node) []*Node {
	var neighbors []*Node
	directions := []data.Position{
		{0, 1},   // Down
		{1, 0},   // Right
		{0, -1},  // Up
		{-1, 0},  // Left
		{1, 1},   // Down-Right (Southeast)
		{-1, 1},  // Down-Left (Southwest)
		{1, -1},  // Up-Right (Northeast)
		{-1, -1}, // Up-Left (Northwest)
	}

	for _, d := range directions {
		newPosition := data.Position{X: node.X + d.X, Y: node.Y + d.Y}
		if newPosition.X >= 0 && newPosition.X < grid.Width && newPosition.Y >= 0 && newPosition.Y < grid.Height && grid.CollisionGrid[newPosition.Y][newPosition.X] {
			neighbors = append(neighbors, &Node{Position: newPosition})
		}
	}

	return neighbors
}

func heuristic(p1, p2 data.Position) float64 {
	dx := math.Abs(float64(p1.X - p2.X))
	dy := math.Abs(float64(p1.Y - p2.Y))
	return math.Max(dx, dy)
}

func reconstructPath(node *Node) []data.Position {
	var path []data.Position
	for node != nil {
		path = append([]data.Position{node.Position}, path...)
		node = node.Parent
	}
	return path
}
