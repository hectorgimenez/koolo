package game

import "github.com/hectorgimenez/d2go/pkg/data"

const (
	CollisionTypeNonWalkable CollisionType = iota
	CollisionTypeWalkable
	CollisionTypeLowPriority
	CollisionTypeMonster
	CollisionTypeObject
)

type CollisionType uint8

type Grid struct {
	OffsetX       int
	OffsetY       int
	Width         int
	Height        int
	CollisionGrid [][]CollisionType
}

func NewGrid(rawCollisionGrid [][]CollisionType, offsetX, offsetY int) *Grid {
	grid := &Grid{
		OffsetX:       offsetX,
		OffsetY:       offsetY,
		Width:         len(rawCollisionGrid[0]),
		Height:        len(rawCollisionGrid),
		CollisionGrid: rawCollisionGrid,
	}

	// Let's lower the priority for the walkable tiles that are close to non-walkable tiles, so we can avoid walking too close to walls and obstacles
	for y := 0; y < len(rawCollisionGrid); y++ {
		for x := 0; x < len(rawCollisionGrid[y]); x++ {
			if rawCollisionGrid[y][x] == CollisionTypeNonWalkable {
				for i := -2; i <= 2; i++ {
					for j := -2; j <= 2; j++ {
						if i == 0 && j == 0 {
							continue
						}
						if y+i < 0 || y+i >= len(rawCollisionGrid) || x+j < 0 || x+j >= len(rawCollisionGrid[y]) {
							continue
						}
						if rawCollisionGrid[y+i][x+j] == CollisionTypeWalkable {
							rawCollisionGrid[y+i][x+j] = CollisionTypeLowPriority
						}
					}
				}
			}
		}
	}

	return grid
}

func (g *Grid) RelativePosition(p data.Position) data.Position {
	return data.Position{
		X: p.X - g.OffsetX,
		Y: p.Y - g.OffsetY,
	}
}

func (g *Grid) IsWalkable(p data.Position) bool {
	p = g.RelativePosition(p)
	return p.X >= 0 && p.X < g.Width && p.Y >= 0 && p.Y < g.Height && g.CollisionGrid[p.Y][p.X] != CollisionTypeNonWalkable
}

func (g *Grid) Copy() *Grid {
	cg := make([][]CollisionType, g.Height)
	for y := 0; y < g.Height; y++ {
		cg[y] = make([]CollisionType, g.Width)
		copy(cg[y], g.CollisionGrid[y])
	}

	return &Grid{
		OffsetX:       g.OffsetX,
		OffsetY:       g.OffsetY,
		Width:         g.Width,
		Height:        g.Height,
		CollisionGrid: cg,
	}
}
