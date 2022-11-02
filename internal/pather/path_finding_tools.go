package pather

import (
	"fmt"
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/koolo/internal/game/area"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

// Kind* constants refer to tile kinds for input and output.
const (
	// KindPlain (.) is a plain tile with a movement cost of 1.
	KindPlain = iota
	// KindBlocker (X) is a tile which blocks movement.
	KindBlocker
	// KindFrom (F) is a tile which marks where the path should be calculated
	// from.
	KindFrom
	// KindTo (T) is a tile which marks the goal of the path.
	KindTo
	// KindPath (●) is a tile to represent where the path is in the output.
	KindPath
	// KindSoftBlocker (S) is a tile to fake some blocking areas with increased cost of moving, like Arcane Sanctuary platforms
	KindSoftBlocker
)

// KindCosts map tile kinds to movement costs.
var KindCosts = map[int]float64{
	KindPlain:       1.0,
	KindFrom:        1.0,
	KindTo:          1.0,
	KindSoftBlocker: 3.0,
}

// A Tile is a tile in a grid which implements Pather.
type Tile struct {
	// Kind is the kind of tile, potentially affecting movement.
	Kind int
	// X and Y are the coordinates of the tile.
	X, Y int
	// W is a reference to the World that the tile is a part of.
	W World
}

// PathNeighbors returns the neighbors of the tile, excluding blockers and
// tiles off the edge of the board.
func (t *Tile) PathNeighbors() []astar.Pather {
	var neighbors []astar.Pather
	if t == nil {
		return neighbors
	}
	for _, offset := range [][]int{
		{-1, 0},
		{1, 0},
		{0, -1},
		{0, 1},
	} {
		if n := t.W.Tile(t.X+offset[0], t.Y+offset[1]); n != nil &&
			n.Kind != KindBlocker {
			neighbors = append(neighbors, n)
		}
	}
	return neighbors
}

// PathNeighborCost returns the movement cost of the directly neighboring tile.
func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	toT := to.(*Tile)
	return KindCosts[toT.Kind]
}

// PathEstimatedCost uses Manhattan distance to estimate orthogonal distance
// between non-adjacent nodes.
func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	toT := to.(*Tile)
	absX := toT.X - t.X
	if absX < 0 {
		absX = -absX
	}
	absY := toT.Y - t.Y
	if absY < 0 {
		absY = -absY
	}
	return float64(absX + absY)
}

// World is a two dimensional map of Tiles.
type World map[int]map[int]*Tile

// Tile gets the tile at the given coordinates in the world.
func (w World) Tile(x, y int) *Tile {
	if w[x] == nil {
		return nil
	}
	return w[x][y]
}

// SetTile sets a tile at the given coordinates in the world.
func (w World) SetTile(t *Tile, x, y int) {
	if w[x] == nil {
		w[x] = map[int]*Tile{}
	}
	w[x][y] = t
	t.X = x
	t.Y = y
	t.W = w
}

// FirstOfKind gets the first tile on the board of a kind, used to get the from
// and to tiles as there should only be one of each.
func (w World) FirstOfKind(kind int) *Tile {
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

// RenderPathImg renders a path on top of a world.
func (w World) RenderPathImg(path []astar.Pather) {
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
				if t.Kind == KindPlain {
					img.Set(x, y, color.White)
				} else {
					img.Set(x, y, color.Black)
				}
			}
		}
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

// ParseWorld parses a textual representation of a world into a world map.
func ParseWorld(collisionGrid [][]bool, fromX, fromY, toX, toY int, ar area.Area) World {
	w := World{}

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
			w.SetTile(&Tile{
				Kind: kind,
			}, y, x)
		}
	}

	w.SetTile(&Tile{
		Kind: KindFrom,
	}, fromX, fromY)

	w.SetTile(&Tile{
		Kind: KindTo,
	}, toX, toY)

	// Hacky solution, sometimes when the character or destination are near a wall pather is not able to calculate
	// the path, so we fake some points around the character making them walkable even if they're not technically
	for i := -2; i < 3; i++ {
		for k := -2; k < 3; k++ {
			if i == 0 && k == 0 {
				continue
			}

			w.SetTile(&Tile{
				Kind: KindPlain,
			}, fromX+i, fromY+k)

			w.SetTile(&Tile{
				Kind: KindPlain,
			}, toX+i, toY+k)
		}
	}

	// Debug only, this will render a png file with map and origin/destination points
	//w.RenderPathImg(nil)
	return w
}
