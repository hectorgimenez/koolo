package pather

import (
	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/hid"
	"math"
	"math/rand"
)

func GetPath(d data.Data, to data.Position, blacklistedCoords ...[2]int) (path *Pather, distance float64, found bool) {
	// Convert to relative coordinates (Current player position)
	fromX := d.PlayerUnit.Position.X - d.AreaOrigin.X
	fromY := d.PlayerUnit.Position.Y - d.AreaOrigin.Y

	// Convert to relative coordinates (Target position)
	toX := to.X - d.AreaOrigin.X
	toY := to.Y - d.AreaOrigin.Y

	// Origin and destination are the same point
	if fromX == toX && fromY == toY {
		return nil, 0, true
	}

	collisionGrid := d.CollisionGrid
	for _, cord := range blacklistedCoords {
		collisionGrid[cord[1]][cord[0]] = false
	}

	w := parseWorld(collisionGrid, fromX, fromY, toX, toY, d.PlayerUnit.Area)

	p, distance, found := astar.Path(w.From(), w.To())

	// Hacky solution, sometimes when the character or destination are near a wall pather is not able to calculate
	// the path, so we fake some points around the character making them walkable even if they're not technically
	if !found && len(blacklistedCoords) == 0 {
		for i := -2; i < 3; i++ {
			for k := -2; k < 3; k++ {
				if i == 0 && k == 0 {
					continue
				}

				w.SetTile(w.NewTile(KindPlain, fromX+i, fromY+k))
				w.SetTile(w.NewTile(KindPlain, toX+i, toY+k))
			}
		}
		p, distance, found = astar.Path(w.From(), w.To())
	}

	// Debug only, this will render a png file with map and origin/destination points
	if config.Config.Debug.RenderMap {
		w.renderPathImg(d, p)
	}

	return &Pather{AstarPather: p, Destination: data.Position{
		X: w.To().X + d.AreaOrigin.X,
		Y: w.To().Y + d.AreaOrigin.Y,
	}}, distance, found
}

func GetClosestWalkablePath(d data.Data, dest data.Position, blacklistedCoords ...[2]int) (path *Pather, distance float64, found bool) {
	maxRange := 20
	step := 4
	dst := 1

	for dst < maxRange {
		for i := -dst; i < dst; i += 1 {
			for j := -dst; j < dst; j += 1 {
				if math.Abs(float64(i)) >= math.Abs(float64(dst)) || math.Abs(float64(j)) >= math.Abs(float64(dst)) {
					cgY := dest.Y - d.AreaOrigin.Y + j
					cgX := dest.X - d.AreaOrigin.X + i
					if len(d.CollisionGrid) > cgY && len(d.CollisionGrid[cgY]) > cgX && d.CollisionGrid[cgY][cgX] {
						return GetPath(d, data.Position{
							X: dest.X + i,
							Y: dest.Y + j,
						}, blacklistedCoords...)
					}
				}
			}
		}
		dst += step
	}

	return nil, 0, false
}

func MoveThroughPath(p *Pather, distance int, teleport bool) {
	moveTo := p.AstarPather[0].(*Tile)
	if distance > 0 && len(p.AstarPather) > distance {
		moveTo = p.AstarPather[len(p.AstarPather)-distance].(*Tile)
	}

	screenX, screenY := GameCoordsToScreenCords(p.AstarPather[len(p.AstarPather)-1].(*Tile).X, p.AstarPather[len(p.AstarPather)-1].(*Tile).Y, moveTo.X, moveTo.Y)
	// Prevent mouse overlap the HUD
	if screenY > int(float32(hid.GameAreaSizeY)/1.21) {
		screenY = int(float32(hid.GameAreaSizeY) / 1.21)
	}

	hid.MovePointer(screenX, screenY)
	if distance > 0 {
		if teleport {
			hid.Click(hid.RightButton)
		} else {
			hid.PressKey(config.Config.Bindings.ForceMove)
		}
	}
}

func GameCoordsToScreenCords(playerX, playerY, destinationX, destinationY int) (int, int) {
	// Calculate diff between current player position and destination
	diffX := destinationX - playerX
	diffY := destinationY - playerY

	// Transform cartesian movement (world) to isometric (screen)e
	// Helpful documentation: https://clintbellanger.net/articles/isometric_math/
	screenX := int((float32(diffX-diffY) * 19.8) + float32(hid.GameAreaSizeX/2))
	screenY := int((float32(diffX+diffY) * 9.9) + float32(hid.GameAreaSizeY/2))

	return screenX, screenY
}

func DistanceFromMe(d data.Data, p data.Position) int {
	return DistanceFromPoint(d.PlayerUnit.Position, p)
}

func DistanceFromPoint(from data.Position, to data.Position) int {
	first := math.Pow(float64(to.X-from.X), 2)
	second := math.Pow(float64(to.Y-from.Y), 2)

	return int(math.Sqrt(first + second))
}

func RandomMovement() {
	midGameX := hid.GameAreaSizeX / 2
	midGameY := hid.GameAreaSizeY / 2
	x := midGameX + rand.Intn(midGameX) - (midGameX / 2)
	y := midGameY + rand.Intn(midGameY) - (midGameY / 2)
	hid.MovePointer(x, y)
	hid.PressKey(config.Config.Bindings.ForceMove)
}
