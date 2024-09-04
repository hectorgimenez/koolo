package astar

import (
	"encoding/gob"
	"os"
	"testing"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

func BenchmarkAstar(b *testing.B) {
	grid := loadGrid()

	start := data.Position{X: 273, Y: 458}
	goal := data.Position{X: 11, Y: 90}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculatePath(grid, start, goal)
	}
}

func TestAstar(t *testing.T) {
	grid := loadGrid()

	start := data.Position{X: 273, Y: 458}
	goal := data.Position{X: 11, Y: 90}

	p, dist, found := CalculatePath(grid, start, goal)
	if int(dist) != 509 {
		t.Errorf("Expected distance to be 509, got %f", dist)
	}
	if len(p) != 510 {
		t.Errorf("Expected path length to be 510, got %d", len(p))
	}
	if !found {
		t.Errorf("Expected path to be found")
	}
}

func loadGrid() *game.Grid {
	var grid game.Grid
	file, err := os.Open("durance_of_hate_grid.bin")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&grid); err != nil {
		panic(err)
	}

	return &grid
}
