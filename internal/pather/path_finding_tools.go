package pather

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
)

// World is a two dimensional map of Tiles.
//type World map[int]map[int]*Tile

type World [][]*Tile

// Tile gets the tile at the given coordinates in the world.
func (w World) Tile(x, y int) *Tile {
	if x < 0 || x > len(w)-1 || w[x] == nil || y > len(w[x])-1 || y < 0 {
		return nil
	}

	return w[x][y]
}

// SetTile sets a tile at the given coordinates in the world.
func (w World) SetTile(t *Tile) {
	w[t.X][t.Y] = t
}

func (w World) NewTile(kind uint8, x, y int) *Tile {
	return &Tile{
		Kind: kind,
		X:    x,
		Y:    y,
		W:    w,
	}
}

// FirstOfKind gets the first tile on the board of a kind, used to get the from
// and to tiles as there should only be one of each.
func (w World) FirstOfKind(kind uint8) *Tile {
	for _, row := range w {
		for _, t := range row {
			if t.Kind == kind {
				return t
			}
		}
	}
	return nil
}

// From gets the from tile from the world.
func (w World) From() *Tile {
	return w.FirstOfKind(KindFrom)
}

// To gets the to tile from the world.
func (w World) To() *Tile {
	return w.FirstOfKind(KindTo)
}

// parseWorld parses a textual representation of a world into a world map.
func parseWorld(expandedGrid bool, collisionGrid [][]bool, ar area.Area) World {
	gridSizeX := len(collisionGrid[0])
	gridSizeY := len(collisionGrid)

	if expandedGrid {
		gridSizeX = expandedGridPadding
		gridSizeY = expandedGridPadding
	}

	w := make(World, gridSizeX)

	for x := 0; x < gridSizeX; x++ {
		w[x] = make([]*Tile, gridSizeY)
	}

	if expandedGrid {
		for x := 0; x < gridSizeY; x++ {
			for y := 0; y < gridSizeX; y++ {
				if x >= 1500 && x <= 1500+len(collisionGrid)-1 && y >= 1500 && y < 1500+len(collisionGrid[0])-1 {
					if collisionGrid[x-1500][y-1500] {
						w.SetTile(w.NewTile(KindPlain, y, x))
					} else {
						w.SetTile(w.NewTile(KindBlocker, y, x))
					}
				} else {
					w.SetTile(w.NewTile(KindPlain, y, x))
				}
			}
		}
	} else {
		for x, xValues := range collisionGrid {
			for y, walkable := range xValues {
				kind := KindBlocker

				// Hacky solution to avoid Arcane Sanctuary A* errors
				if ar == area.ArcaneSanctuary {
					kind = KindSoftBlocker
				}

				if walkable {
					kind = KindPlain
				}

				w.SetTile(w.NewTile(kind, y, x))
			}
		}
	}

	return w
}

// RenderPathImg renders a path on top of a world.
func (w World) renderPathImg(d data.Data, path []astar.Pather, expandedCG bool) {
	width := len(w)
	if width == 0 {
		return
	}
	height := len(w[0])

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), img, image.Point{}, draw.Over)

	pathLocs := map[string]bool{}
	for _, p := range path {
		pT := p.(*Tile)
		pathLocs[fmt.Sprintf("%d,%d", pT.X, pT.Y)] = true
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			t := w.Tile(x, y)
			if pathLocs[fmt.Sprintf("%d,%d", x, y)] {
				img.Set(x, y, color.RGBA{
					R: 36,
					G: 255,
					B: 0,
					A: 255,
				})
			} else if t != nil {
				switch t.Kind {
				case KindPlain:
					img.Set(x, y, color.White)
				case KindBlocker:
					img.Set(x, y, color.Black)
				case KindSoftBlocker:
					img.Set(x, y, color.RGBA{238, 238, 238, 255})
				}
			}
		}
	}

	for _, r := range d.Rooms {
		rPosX, rPosY := relativePosition(d, r.GetCenter(), expandedCG)
		img.Set(rPosX, rPosY, color.RGBA{204, 204, 0, 255})
	}

	img.Set(w.From().X, w.From().Y, color.RGBA{
		R: 255, G: 0, B: 0, A: 255,
	})

	img.Set(w.To().X, w.To().Y, color.RGBA{
		R: 0, G: 0, B: 255, A: 255,
	})

	for _, o := range d.Objects {
		oPosX, oPosY := relativePosition(d, o.Position, expandedCG)
		img.Set(oPosX, oPosY, color.RGBA{255, 165, 0, 255})
	}

	for _, m := range d.Monsters {
		mPosX, mPosY := relativePosition(d, m.Position, expandedCG)
		img.Set(mPosX, mPosY, color.RGBA{255, 0, 255, 255})
	}

	outFile, _ := os.Create("cg.png")
	defer outFile.Close()
	png.Encode(outFile, img)
}
