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
	"github.com/hectorgimenez/koolo/internal/game"
)

type World struct {
	World [][]*Tile
	from  data.Position
	to    data.Position
}

// Tile gets the tile at the given coordinates in the World.
func (w *World) Tile(x, y int) *Tile {
	if x < 0 || x > len(w.World)-1 || w.World[x] == nil || y > len(w.World[x])-1 || y < 0 {
		return nil
	}

	return w.World[x][y]
}

// SetTile sets a tile at the given coordinates in the World, if the World is big enough.
func (w *World) SetTile(t *Tile) {
	if len(w.World) >= t.X+1 && len(w.World[t.X]) >= t.Y+1 {
		w.World[t.X][t.Y] = t
	}
}

func (w *World) NewTile(kind uint8, x, y int) *Tile {
	walkable := false
	cost := 0.0
	if kind != KindBlocker {
		walkable = true
		cost = KindCosts[kind]
	}

	return &Tile{
		Kind:     kind,
		Walkable: walkable,
		Cost:     cost,
		X:        x,
		Y:        y,
		W:        w,
	}
}

func (w *World) From() *Tile {
	return w.World[w.from.X][w.from.Y]
}

func (w *World) To() *Tile {
	return w.World[w.to.X][w.to.Y]
}

func (w *World) SetFrom(position data.Position) {
	w.from = position
}
func (w *World) SetTo(position data.Position) {
	w.to = position
}

// RenderPathImg renders a path on top of a World.
func (w *World) renderPathImg(d *game.Data, path []astar.Pather, cgOffset data.Position) {
	width := len(w.World)
	if width == 0 {
		return
	}
	height := len(w.World[0])

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
		rPosX, rPosY := relativePosition(d, r.GetCenter(), cgOffset)
		img.Set(rPosX, rPosY, color.RGBA{204, 204, 0, 255})
	}

	for _, o := range d.Objects {
		oPosX, oPosY := relativePosition(d, o.Position, cgOffset)
		if o.IsDoor() {
			img.Set(oPosX, oPosY, color.RGBA{101, 67, 33, 255})
		} else {
			img.Set(oPosX, oPosY, color.RGBA{255, 165, 0, 255})
		}
	}

	for _, m := range d.Monsters {
		mPosX, mPosY := relativePosition(d, m.Position, cgOffset)
		img.Set(mPosX, mPosY, color.RGBA{255, 0, 255, 255})
	}

	img.Set(w.From().X, w.From().Y, color.RGBA{
		R: 255, G: 0, B: 0, A: 255,
	})

	img.Set(w.To().X, w.To().Y, color.RGBA{
		R: 0, G: 0, B: 255, A: 255,
	})

	outFile, _ := os.Create("cg.png")
	defer outFile.Close()
	png.Encode(outFile, img)
}
