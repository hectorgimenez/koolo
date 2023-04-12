package pather

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/d2go/pkg/data"
)

type Pather struct {
	AstarPather
	Destination data.Position
}

type AstarPather []astar.Pather

func (p *Pather) Distance() int {
	return len(p.AstarPather)
}
