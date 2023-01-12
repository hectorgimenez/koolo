package pather

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Pather struct {
	AstarPather
	Destination game.Position
}

type AstarPather []astar.Pather

func (p *Pather) Distance() int {
	return len(p.AstarPather)
}
