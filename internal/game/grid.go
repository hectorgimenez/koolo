package game

import "github.com/hectorgimenez/d2go/pkg/data"

type Grid struct {
	OffsetX       int
	OffsetY       int
	Width         int
	Height        int
	CollisionGrid [][]bool
}

func NewGrid(collisionGrid [][]bool, offsetX, offsetY int) *Grid {
	return &Grid{
		OffsetX:       offsetX,
		OffsetY:       offsetY,
		Width:         len(collisionGrid[0]),
		Height:        len(collisionGrid),
		CollisionGrid: collisionGrid,
	}
}

func (g *Grid) RelativePosition(p data.Position) data.Position {
	return data.Position{
		X: p.X - g.OffsetX,
		Y: p.Y - g.OffsetY,
	}
}

func (g *Grid) IsWalkable(p data.Position) bool {
	p = g.RelativePosition(p)
	return p.X >= 0 && p.X < g.Width && p.Y >= 0 && p.Y < g.Height && g.CollisionGrid[p.Y][p.X]
}
