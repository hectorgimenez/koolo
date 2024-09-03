package astar

import (
	"testing"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

func BenchmarkAstar(b *testing.B) {
	grid := game.NewGrid(collisionGrid, 0, 0)

	start := data.Position{X: 39, Y: 441}
	goal := data.Position{X: 264, Y: 158}

	b.ResetTimer()
	CalculatePath(grid, start, goal)
}

func TestAstar(t *testing.T) {
	grid := game.NewGrid(collisionGrid, 0, 0)

	start := data.Position{X: 39, Y: 441}
	goal := data.Position{X: 264, Y: 158}

	p, dist, found := CalculatePath(grid, start, goal)
	if dist != 514 {
		t.Errorf("Expected distance to be 514, got %f", dist)
	}
	if len(p) != 515 {
		t.Errorf("Expected path length to be 514, got %d", len(p))
	}
	if !found {
		t.Errorf("Expected path to be found")
	}
}
