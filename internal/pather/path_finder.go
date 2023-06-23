package pather

import (
	"math"
	"math/rand"

	"github.com/beefsack/go-astar"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/hid"
)

const expandedGridPadding = 3000

func GetPath(d data.Data, to data.Position, blacklistedCoords ...[2]int) (path *Pather, distance float64, found bool) {
	expandedCG := shouldExpandCollisionGrid(d, to)
	// Convert to relative coordinates (Current player position)
	fromX, fromY := relativePosition(d, d.PlayerUnit.Position, expandedCG)

	// Convert to relative coordinates (Target position)
	toX, toY := relativePosition(d, to, expandedCG)

	// Origin and destination are the same point
	if fromX == toX && fromY == toY {
		return nil, 0, true
	}

	collisionGrid := d.CollisionGrid

	// Lut Gholein map is a bit bugged, we should close this fake path to avoid pathing issues
	if d.PlayerUnit.Area == area.LutGholein {
		collisionGrid[13][210] = false
	}

	for _, cord := range blacklistedCoords {
		collisionGrid[cord[1]][cord[0]] = false
	}

	// Add some padding to the origin/destination, sometimes when the origin or destination are close to a non-walkable
	// area, pather is not able to calculate the path, so we add some padding around origin/dest to avoid this
	for i := -3; i < 4; i++ {
		for k := -3; k < 4; k++ {
			if i == 0 && k == 0 {
				continue
			}

			//collisionGrid[ensureValueInCG(fromY+i, len(collisionGrid))][ensureValueInCG(fromX+k, len(collisionGrid[0]))] = true
			collisionGrid[ensureValueInCG(toY+i, len(collisionGrid))][ensureValueInCG(toX+k, len(collisionGrid[0]))] = true
		}
	}

	w := parseWorld(expandedCG, collisionGrid, d.PlayerUnit.Area)

	// Set Origin and Destination points
	w.SetTile(w.NewTile(KindFrom, fromX, fromY))
	w.SetTile(w.NewTile(KindTo, toX, toY))

	p, distance, found := astar.Path(w.From(), w.To())

	if config.Config.Debug.RenderMap {
		w.renderPathImg(d, p, expandedCG)
	}

	// Debug only, this will render a png file with map and origin/destination points
	if config.Config.Debug.RenderMap {
		w.renderPathImg(d, p, expandedCG)
	}

	return &Pather{AstarPather: p, Destination: data.Position{
		X: w.To().X + d.AreaOrigin.X,
		Y: w.To().Y + d.AreaOrigin.Y,
	}}, distance, found
}

func ensureValueInCG(val, cgSize int) int {
	if val < 0 {
		return 0
	}

	if val >= cgSize {
		return cgSize - 1
	}

	return val
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
					if cgX > 0 && cgY > 0 && len(d.CollisionGrid) > cgY && len(d.CollisionGrid[cgY]) > cgX && d.CollisionGrid[cgY][cgX] {
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
	if len(p.AstarPather) == 0 {
		if teleport {
			hid.Click(hid.RightButton)
		} else {
			hid.PressKey(config.Config.Bindings.ForceMove)
		}
		return
	}

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

func relativePosition(d data.Data, p data.Position, expandedCG bool) (int, int) {
	if expandedCG {
		return p.X - d.AreaOrigin.X + expandedGridPadding/2, p.Y - d.AreaOrigin.Y + expandedGridPadding/2
	}

	return p.X - d.AreaOrigin.X, p.Y - d.AreaOrigin.Y
}

func shouldExpandCollisionGrid(d data.Data, p data.Position) bool {
	relativeToX := p.X - d.AreaOrigin.X
	relativeToY := p.Y - d.AreaOrigin.Y

	return relativeToX < 0 || relativeToY < 0 || relativeToX > len(d.CollisionGrid[0]) || relativeToY > len(d.CollisionGrid)
}
