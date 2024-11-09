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
				switch grid.CollisionGrid[y][x] {
				case game.CollisionTypeNonWalkable:
					img.Set(x, y, color.Black)
				case game.CollisionTypeWalkable:
					img.Set(x, y, color.White)
				case game.CollisionTypeLowPriority:
					img.Set(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255}) // Gray
				case game.CollisionTypeMonster:
					img.Set(x, y, color.RGBA{R: 255, A: 255}) // Red
				case game.CollisionTypeObject:
					img.Set(x, y, color.RGBA{R: 160, G: 32, B: 240, A: 255}) // Purple
				}
			}
		}
	}

	for _, r := range pf.data.Rooms {
		pos := grid.RelativePosition(r.GetCenter())
		img.Set(pos.X, pos.Y, color.RGBA{R: 204, G: 204, A: 255}) // Dark yellow
	}

	img.Set(from.X, from.Y, color.RGBA{R: 158, G: 0, B: 0, A: 255}) // Garnet

	img.Set(to.X, to.Y, color.RGBA{R: 0, G: 0, B: 255, A: 255}) // Blue

	outFile, _ := os.Create("cg.png")
	defer outFile.Close()
	png.Encode(outFile, img)
}
