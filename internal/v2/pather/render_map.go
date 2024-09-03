package pather

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/game"
)

func (pf *PathFinder) renderMap(grid *game.Grid, from, to data.Position, path Path) {
	img := image.NewRGBA(image.Rect(0, 0, grid.Width, grid.Height))
	draw.Draw(img, img.Bounds(), img, image.Point{}, draw.Over)

	pathLocs := map[string]bool{}
	for _, p := range path {
		pathLocs[fmt.Sprintf("%d,%d", p.X, p.Y)] = true
	}

	for x := 0; x < grid.Width; x++ {
		for y := 0; y < grid.Height; y++ {
			if pathLocs[fmt.Sprintf("%d,%d", x, y)] {
				img.Set(x, y, color.RGBA{R: 36, G: 255, B: 0, A: 255})
			} else {
				if grid.CollisionGrid[y][x] {
					img.Set(x, y, color.White)
				} else {
					img.Set(x, y, color.Black)
				}
			}
		}
	}

	for _, r := range pf.data.Rooms {
		pos := grid.RelativePosition(r.GetCenter())
		img.Set(pos.X, pos.Y, color.RGBA{R: 204, G: 204, A: 255})
	}

	for _, o := range pf.data.Objects {
		pos := grid.RelativePosition(o.Position)
		if o.IsDoor() {
			img.Set(pos.X, pos.Y, color.RGBA{R: 101, G: 67, B: 33, A: 255})
		} else {
			img.Set(pos.X, pos.Y, color.RGBA{R: 255, G: 165, A: 255})
		}
	}

	for _, m := range pf.data.Monsters {
		pos := grid.RelativePosition(m.Position)
		img.Set(pos.X, pos.Y, color.RGBA{R: 255, B: 255, A: 255})
	}

	img.Set(from.X, from.Y, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	img.Set(to.X, to.Y, color.RGBA{R: 0, G: 0, B: 255, A: 255})

	outFile, _ := os.Create("cg.png")
	defer outFile.Close()
	png.Encode(outFile, img)
}
